import importlib
import json
import os

from flask import Flask, jsonify, request

app = Flask(__name__)

files = list(filter(lambda file: file.endswith(".py"), os.listdir("./customLogic")))
print("found files %s" % files)

def getHandler(fileNoExt):
    module = importlib.import_module("." + fileNoExt, package="customLogic")
    attrs = map(lambda v: getattr(module, v), filter(lambda v: not v.startswith("__"), vars(module)))
    customLogic = next(filter(lambda attr: callable(attr), attrs))
    def handler():
        output = customLogic(request.get_json())
        return jsonify(output)
    return handler

for file in files:
    fileNoExt = file[:-3]
    app.add_url_rule("/" + fileNoExt, fileNoExt, getHandler(fileNoExt), methods=["POST"])


@app.route("/ping")
def ping():
    return "pong"
