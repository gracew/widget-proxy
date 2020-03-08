import json

from flask import Flask, jsonify, request

app = Flask(__name__)

with open("./customLogic.json") as f:
    custom_logic = json.loads(f.read())
    for el in custom_logic:
        if el["operationType"] == "CREATE":
            if el["beforeSave"]:
                with open("./before_save.py", "w+") as f_out:
                    f_out.write(el["beforeSave"])

                @app.route("/beforeSave", methods=["POST"])
                def before_save():
                    from before_save import before_save
                    output = before_save(request.get_json())
                    return jsonify(output)

            if el["afterSave"]:
                with open("./after_save.py", "w+") as f_out:
                    f_out.write(el["afterSave"])

                @app.route("/afterSave", methods=["POST"])
                def after_save():
                    from after_save import after_save
                    output = after_save(request.get_json())
                    return jsonify(output)
