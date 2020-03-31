import json
import os

from flask import Flask, jsonify, request

app = Flask(__name__)

BEFORE_CREATE = "./customLogic/beforeCreate.py"
AFTER_CREATE = "./customLogic/afterCreate.py"

if os.path.exists(BEFORE_CREATE):
    @app.route("/beforeCreate", methods=["POST"])
    def beforeCreate():
        from customLogic.beforeCreate import beforeCreate
        output = beforeCreate(request.get_json())
        return jsonify(output)

if os.path.exists(AFTER_CREATE):
    @app.route("/afterCreate", methods=["POST"])
    def afterCreate():
        from customLogic.afterCreate import afterCreate
        output = afterCreate(request.get_json())
        return jsonify(output)


@app.route("/ping")
def ping():
    return
