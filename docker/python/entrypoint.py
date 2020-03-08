import json

from custom_logic import before_save

with open("./input.json") as f_in:
    input = json.loads(f_in.read())
    output = before_save(input)
    with open("./output/output.json", "w+") as f_out:
        f_out.write(json.dumps(output))