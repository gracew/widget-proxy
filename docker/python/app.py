import importlib
import json
import os

from flask import Flask, jsonify, request

app = Flask(__name__)

customLogicDir = "./customLogic/"
files = list(filter(lambda file: file.endswith(".py"), os.listdir(customLogicDir)))
print("found files %s" % files)

for file in files:
    fileNoExt = file[:-3]
    module = importlib.import_module("." + fileNoExt, package="customLogic")
    attrs = map(lambda v: getattr(module, v), filter(lambda v: not v.startswith("__"), vars(module)))
    customLogic = next(filter(lambda attr: callable(attr), attrs))
    @app.route("/" + fileNoExt, methods=["POST"])
    def handler():
        output = customLogic(request.get_json())
        return jsonify(output)
    handler.__name__ = fileNoExt


@app.route("/ping")
def ping():
    return "pong"
