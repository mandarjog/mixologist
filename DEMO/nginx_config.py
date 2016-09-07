import urlparse
import sys
import cStringIO
import re
import socket



# This is somewhat nginx.conf specific, but could be generalized
namesre = re.compile("(.*)(http(s){0,1})://(\S*)(\s*);(.*)")


class DNSLine(object):
    """
    Store and Process config lines with dns entries
    """

    def __init__(self, prefix, url, suffix):
        self.prefix = prefix
        self.url = url
        self.suffix = suffix

    def eval(self, names):
        newurl = self.url._replace(netloc="{}:{}".format(
            names[self.url.hostname],
            self.url.port)).geturl()
        return "{}{}{}".format(
            self.prefix,
            newurl,
            self.suffix)

class DNSResolver(object):
    def resolve(self, name):
        try:
            return socket.gethostbyname(name)
        except socket.gaierror as se:
            print se
            return name

class Cfg(object):

    def __init__(self):
        self.lines = []
        self.names = {}
        self.resolver = DNSResolver()

    def add(self, line):
        self.lines.append(line)

    def addDNS(self, line):
        mm = namesre.match(line)
        # 1, 2, 4 ,5 contact = line
        if mm:
            url = urlparse.urlparse(mm.group(2) + "://" + mm.group(4))
            if url.hostname and url.hostname.upper() != url.hostname.lower():
                self.names[url.hostname] = ""
                dnsLine = DNSLine(
                    mm.group(1),
                    url,
                    mm.group(5) + ";" + mm.group(6))
                self.lines.append(dnsLine)
            else:
                self.lines.append(line)
        else:
            self.lines.append(line)

    def resolve(self, resolver):
        """
        True if any of the names to be resolved have changed
        """
        resolver = resolver or self.resolver
        newnames = {hn: resolver.resolve(hn) for hn in self.names}
        changed = newnames != self.names
        if changed:
            self.names = newnames

        return changed

    def eval(self):
        out = cStringIO.StringIO()
        for line in self.lines:
            if hasattr(line, 'eval'):
                print >>out, line.eval(self.names)
            else:
                print >>out, line,
        output = out.getvalue()
        out.close()
        return output


def parse_config(filename, start_markers=None):
    """
    parse file and return lines and names
    """
    with open(filename, "rt") as fl:
        return _parse_config(fl, start_markers)


def _parse_config(fileobj, start_markers=None):
    """
    parse file and return lines and names
    """
    start_markers = start_markers or ["service_control", "proxy_pass"]
    cfg = Cfg()

    for line in fileobj:
        if any([line.lstrip().startswith(sm) for sm in start_markers]):
            cfg.addDNS(line)
        else:
            cfg.add(line)
    return cfg


def get_args():
    import argparse
    argp = argparse.ArgumentParser()
    argp.add_argument("--timeout", type=int, default=900, help="timeout in seconds, process will exit if there is no change")
    argp.add_argument("--poll-interval", type=int, default=30, help="poll interval in seconds")
    argp.add_argument("--exit-code-on-change", type=int, default=2, help="Process will exit with this exit code if")
    argp.add_argument("-c", required=True, help="source nginx conf file, conf directory should be writable")
    argp.add_argument("-d", required=True, help="destination nginx conf file, same directory as source nginx.conf")

    return argp


def main(argv):
    args = get_args().parse_args(argv)


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
