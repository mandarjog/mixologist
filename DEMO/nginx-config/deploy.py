#!/usr/bin/env python
import json
import string
import os
import tempfile
import urlparse
import logging

import sh
from sh import kubectl as _kubectl

THIS_DIR = os.path.dirname(os.path.abspath(__file__))


DEPLOY_MAP = {
    "BOOKSTORE_IMAGE": "gcr.io/mixologist-142215/bookstore-mixologist",
    "MIXOLOGIST_IMAGE": "gcr.io/mixologist-142215/mixologist",
    "servicecontrol": "http://mixologist:9092/",
    "ESP_IMAGE": "gcr.io/mixologist-142215/endpoints-runtime"
    # afte esp is realsed the above should change to
    # b.gcr.io/endpoints/endpoints-runtime
}


def get_args():
    import argparse
    argp = argparse.ArgumentParser(
        formatter_class=argparse.ArgumentDefaultsHelpFormatter
    )
    argp.add_argument("--namespace", 
                      default=os.getenv("NAMESPACE"),
                      help="kubernetes namespace or set NAMESPACE=")
    argp.add_argument("--MIXOLOGIST-IMAGE", default=DEPLOY_MAP["MIXOLOGIST_IMAGE"],
                      help="Mixologist Docker image")
    argp.add_argument("--BOOKSTORE-IMAGE", default=DEPLOY_MAP["BOOKSTORE_IMAGE"],
                      help="Boostore Docker image")
    argp.add_argument("--ESP-IMAGE", default=DEPLOY_MAP["ESP_IMAGE"],
                      help="ESP Docker image")
    argp.add_argument("--servicecontrol", default=DEPLOY_MAP["servicecontrol"],
                      help="Service Control Server used by ESP")
    argp.add_argument("--kube-template", help="template for kubenetes service creation",
                      default=THIS_DIR + "/mixologist_bookstore_demo.yml")
    argp.add_argument("--service-json-template", help="template for api service",
                      default=THIS_DIR + "/bookstore.json")
    argp.add_argument("--dns", help="dns server used by ESP. default: kube-dns")
    argp.add_argument("--force-delete-namespace", help="Delete namespace if it already exists", action="store_true", default=False)
    argp.add_argument("-v", type=int, help="verbosity", default=1)
    return argp


def validate_args(parser, args):
    if args.namespace is None:
        parser.error("--namespace is required if NAMESPACE variable is not set")

class KubeCtl(object):

    def __init__(self, namespace, log=None):
        self.namespace = namespace
        self.l = log

    def _cmd_(self, cmd, ns=True, js=True):
        args = [cmd.split()]
        if ns:
            args.append("--namespace=" + self.namespace)
        if js:
            args.append("-o=json")

        proc = _kubectl(*args)
        self.l.info(proc.ran)
        output = proc.stdout

        if js:
            return json.loads(output)
        else:
            return output

    def pods(self):
        return(self._cmd_("get pods"))

    def delete_namespace(self):
        self.l.info("delete_namespace() "+self.namespace)
        return self._cmd_("delete namespace " + self.namespace, js=False)


    def get_namespace(self):
        try:
            return self._cmd_("get namespace " + self.namespace)
        except sh.ErrorReturnCode_1 as ex:
            if 'not found' not in ex.stderr:
                raise
        return None

    def create_namespace(self, force):
        self.l.info("create_namespace() "+self.namespace)
        ns = self.get_namespace()

        if ns is not None:
            if force:
                self.delete_namespace()
            else:
                return ns

        return self._cmd_("create namespace " + self.namespace)

    def create_configmap(self, mapname, filepath, recreate=False):
        self.l.info("create_configmap() {} {} {}".format(self.namespace, mapname, filepath))
        if recreate:
            self.delete_configmap(mapname)

        try:
            return self._cmd_("get configmap " + mapname)
        except sh.ErrorReturnCode_1 as ex:
            if 'not found' not in ex.stderr:
                raise

        return self._cmd_("create configmap " + mapname + " --from-file=" + filepath)

    def delete_configmap(self, mapname):
        try:
            return self._cmd_("delete configmap " + mapname, js=False)
        except sh.ErrorReturnCode_1 as ex:
            if 'not found' not in ex.stderr:
                raise

    ##TODO check why an otherwise valid yml fails validation
    def create(self, ymlfile):
        self.l.info(ymlfile)
        return self._cmd_("create -f " + ymlfile + " --validate=false", js=False)

    def get_cluster_dns(self):
        op = self._cmd_("get svc kube-dns --namespace kube-system", ns=False)
        return op["spec"]["clusterIP"]

    def get_svc_endpoint(self, svc):
        spec = self._cmd_("get svc "+svc)["spec"]
        return "{}:{}".format(
            spec["clusterIP"],
            spec["ports"][0]["port"])

def process_template(inputfile, outputfile, varmap):
    with open(inputfile, 'rt') as fl:
        output = string.Template(fl.read()).substitute(varmap)
        with open(outputfile, 'wt') as wl:
            wl.write(output)


def deploy(args, log):
    kubectl = KubeCtl(args.namespace, log)

    # check / create namespace
    kubectl.create_namespace(force=args.force_delete_namespace)

    # hydrate templates with the given info
    varmap = {k: args.__dict__[k] for k in DEPLOY_MAP}
    varmap["dns"] = args.dns or kubectl.get_cluster_dns()
    
    # This should really work, 
    # working around ESP dns issue
    # ESP is unable to resolve the symbolic name of service control
    # it needs fqdn
    url = urlparse.urlparse(varmap["servicecontrol"])
    if '.' not in url.hostname:
        # ensure that service control uses fqdn
        scheme = url.scheme or "http"
        varmap["servicecontrol"] = \
            "{scheme}://{host}.{namespace}.svc.cluster.local:{port}".format(
                scheme=scheme,
                host=url.hostname,
                namespace=args.namespace,
                port=url.port)

    log.debug("using {}".format(varmap))

    tdir = tempfile.mkdtemp("mixologist-deploy")
    kube_yml = tdir + "/kube.yml"
    service_json = tdir + "/bookstore.json"
    process_template(args.kube_template, kube_yml, varmap)
    process_template(args.service_json_template, service_json, varmap)

    # create config maps
    kubectl.create_configmap("bookstore-service-config",
                             service_json,
                             True)
    kubectl.create_configmap("prometheus-config",
                             THIS_DIR + "/prometheus.yml",
                             True)

    # call "create" with the template
    kubectl.create(kube_yml)

    log.info("Created Services, waiting for deployment to become available")

    return 0


def main(argv):
    argp = get_args()
    args = argp.parse_args(argv)
    FORMAT = '[%(asctime)s] p%(process)s {%(pathname)s:%(lineno)d} %(levelname)s - %(message)s'
    logging.basicConfig(format=FORMAT)
    log = logging.getLogger("deploy")
    log.setLevel(args.v)

    validate_args(argp, args)
    return deploy(args, log)


if __name__ == "__main__":
    import sys
    sys.exit(main(sys.argv[1:]))
