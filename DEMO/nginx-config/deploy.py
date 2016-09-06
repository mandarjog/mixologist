#!/usr/bin/env python
import json
import string
import os
import tempfile

import sh
from sh import kubectl as _kubectl

THIS_DIR = os.path.dirname(os.path.abspath(__file__))


DEPLOY_MAP = {
    "BOOKSTORE_IMAGE": "gcr.io/mixologist-142215/bookstore-mixologist",
    "MIXOLOGIST_IMAGE": "gcr.io/mixologist-142215/mixologist",
    "servicecontrol": "mixologist:9092",
    "ESP_IMAGE": "gcr.io/mixologist-142215/endpoints-runtime"
    # afte esp is realsed the above should change to
    # b.gcr.io/endpoints/endpoints-runtime
}


def get_args():
    import argparse
    argp = argparse.ArgumentParser(
        formatter_class=argparse.ArgumentDefaultsHelpFormatter
    )
    argp.add_argument("--namespace", required=True,
                      help="kubernetes namespace")
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

    return argp


class KubeCtl(object):

    def __init__(self, namespace):
        self.namespace = namespace

    def _cmd_(self, cmd, ns=True, js=True):
        args = [cmd.split()]
        if ns:
            args.append("--namespace=" + self.namespace)
        if js:
            args.append("-o=json")

        output = _kubectl(*args).stdout

        if js:
            return json.loads(output)
        else:
            return output

    def pods(self):
        return(self._cmd_("get pods"))

    def create_namespace(self):
        try:
            return self._cmd_("get namespace " + self.namespace)
        except sh.ErrorReturnCode_1 as ex:
            if 'not found' not in ex.stderr:
                raise

        return self._cmd_("create namespace " + self.namespace)

    def create_configmap(self, mapname, filepath, recreate=False):
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
        return self._cmd_("create -f " + ymlfile + " --validate=false", js=False)


def process_template(inputfile, outputfile, varmap):
    with open(inputfile, 'rt') as fl:
        output = string.Template(fl.read()).substitute(varmap)
        with open(outputfile, 'wt') as wl:
            wl.write(output)


def deploy(args):
    kubectl = KubeCtl(args.namespace)
    # check / create namespace
    kubectl.create_namespace()

    # hydrate templates with the given info
    varmap = {k: args.__dict__[k] for k in DEPLOY_MAP}
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


def main(argv):
    args = get_args().parse_args(argv)
    return deploy(args)


if __name__ == "__main__":
    import sys
    sys.exit(main(sys.argv[1:]))
