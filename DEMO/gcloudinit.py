#!/usr/bin/env python
import os
import logging

import sh
from sh import gcloud as _gcloud 


def get_args():
    import argparse
    argp = argparse.ArgumentParser(
        formatter_class=argparse.ArgumentDefaultsHelpFormatter
    )
    argp.add_argument("--project", 
                      default=os.getenv("PROJECT_ID"),
                      help="gcloud project or set PROJECT_ID=")
    argp.add_argument("--cluster-name", default="mixologist")
    argp.add_argument("-v", type=int, help="verbosity", default=1)

    return argp

def validate_args(parser, args):
    if args.project is None:
        parser.error("--project required if PROJECT_ID is not set")


def get_gcloud(log):
    def gcloud(cmd):
        proc = _gcloud(cmd.split(), _bg=True)
        log.info(proc.ran)
        proc.wait()
    return gcloud

def gcloudinit(args, log):
    gcloud = get_gcloud(log)
    gcloud("components install kubectl")
    gcloud("config set project {}".format(args.project))
    gcloud("config set compute/zone us-central1-b")
    if args.project == 'mixologist-142215' and\
            args.cluster_name == 'mixologist':
        log.info("Skipping cluster creation")
    else:
        gcloud("container clusters create {}".format(args.cluster_name))

    gcloud("config set container/cluster {}".format(args.cluster_name))
    gcloud("container clusters get-credentials {}".format(args.cluster_name))

def main(argv):
    argp = get_args()
    args = argp.parse_args(argv)
    FORMAT = '[%(asctime)s] p%(process)s {%(pathname)s:%(lineno)d} %(levelname)s - %(message)s'
    logging.basicConfig(format=FORMAT)
    log = logging.getLogger("init")
    log.setLevel(args.v)

    validate_args(argp, args)
    return gcloudinit(args, log)


if __name__ == "__main__":
    import sys
    sys.exit(main(sys.argv[1:]))
