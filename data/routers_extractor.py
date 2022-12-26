import yaml
import json
import sys
import inspect
import argparse
import importlib.util

ROUTERS_YML = "./routers.yml"
FIRMWARES = "./firmwares.py"

class Extractor:
    def __init__(self, routers_file):
        self.routers_file = routers_file
        self.routers_data = self.read_routers()
        self.firmwares_data = self.get_firmwares_data()
        self.routers = self.create_router_data()

    def read_routers(self):
        with open(self.routers_file, "r") as fd:
            return yaml.load(fd, Loader=yaml.FullLoader)

    def get_firmwares_data(self):
        firmware_clasess = inspect.getmembers(sys.modules["firmwares"])
        firmware_data = {}

        for name, cls in firmware_clasess:
            try:
                if "ssh" in str(cls.whitebox_data.bases).lower():
                    protocol = "ssh"
                else:
                    protocol = "telnet"

                #name = cls.name
                #if "archer" in cls.name or "vz" in cls.name:
                    #name = cls.name.replace("_", "")

                login_info = cls.whitebox_data.login_info
                if login_info:
                    login_info = list(login_info)

                name = name.lower()

                firmware_data[name] = {
                    "name": name,
                    "busybox": cls.whitebox_data.path.busybox,
                    "sam_dir": cls.whitebox_data.path.sam,
                    "login_info": login_info,
                    "protocol": protocol,
                    "communication_port": str(cls.communication_port)
                }
            except Exception as e:
                pass
        return firmware_data

    def create_router_data(self):
        routers = []
        for name, data in self.routers_data.items():
            try:
                fw = data["fw"]

                routers.append({
                    "name": name,
                    "ip": data["ip"],
                    "ssid": data["wifi"]["2.4"][0],
                    "password": data["wifi"]["2.4"][1],
                    **self.firmwares_data[fw]
                })
            except Exception as e:
                pass
        return routers


def parse_args():
    parser = argparse.ArgumentParser()
    parser.add_argument("-r", "--routers-file", type=str, default=ROUTERS_YML, help="routers.yml path")
    parser.add_argument("-f", "--firmwares-module", type=str, default=FIRMWARES, help="firmwares module path")

    args = parser.parse_args()

    return args.routers_file, args.firmwares_module


def import_firmwares_module(fw_path):
    spec = importlib.util.spec_from_file_location("firmwares", fw_path)
    fw = importlib.util.module_from_spec(spec)
    sys.modules["firmwares"] = fw
    spec.loader.exec_module(fw)


if __name__ == "__main__":
    routers_file, firmwares_module = parse_args()

    import_firmwares_module(firmwares_module)

    ex = Extractor(routers_file)

    with open("/tmp/data.json", "w") as fd:
        fd.write(json.dumps(ex.routers))