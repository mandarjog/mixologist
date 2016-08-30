import nginx_config
import StringIO
import string

TMPL = """
#pid nginx.pid;
daemon off;
error_log stderr info;

events {
    worker_connections  32;
}

http {
    client_body_timeout 600s;
    client_header_timeout 600s;
    proxy_send_timeout 600s;
    proxy_read_timeout 600s;
    client_body_temp_path tmp;
    proxy_temp_path tmp;
    access_log /dev/stdout;

    server {
        listen       8090;

        location / {
            endpoints {
              on;
              api bookstore.json;
              service_control http://%(mixologist)s:9092/;
            }
            proxy_pass http://%(bookstore)s:8080/;
        }
    }
}
"""


class MapResolver(object):

    def __init__(self, mapping):
        self.mapping = mapping

    def resolve(self, name):
        return self.mapping.get(name)


def test_cfg_parse1():
    inputstr = TMPL % {"mixologist": "mixologist", "bookstore": "bookstore"}
    cfg = nginx_config._parse_config(StringIO.StringIO(inputstr))
    mapping = {"mixologist": "10.10.11.1", "bookstore": "10.10.10.1"}
    resolver = MapResolver(mapping)

    assert cfg.resolve(resolver)

    # print TMPL % mapping
    # print cfg.eval()

    assert cfg.eval() == TMPL % mapping


def test_cfg_parse2():
    inputstr = TMPL % {"mixologist": "mixologist.mix",
                       "bookstore": "bookstore.mix"}
    cfg = nginx_config._parse_config(StringIO.StringIO(inputstr))
    mapping = {"mixologist.mix": "10.10.11.1", "bookstore.mix": "10.10.10.1"}
    mapping_template = {"mixologist": "10.10.11.1", "bookstore": "10.10.10.1"}
    resolver = MapResolver(mapping)

    assert cfg.resolve(resolver)

    # print TMPL % mapping_template
    # print cfg.eval()

    assert cfg.eval() == TMPL % mapping_template
