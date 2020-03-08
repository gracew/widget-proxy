const fs = require("fs");

const express = require("express");
const bodyParser = require("body-parser");

const app = express();
const port = 8080;

app.use(bodyParser.json());

const customLogic = JSON.parse(fs.readFileSync("./customLogic.json"));
customLogic.forEach(el => {
  if (el.operationType === "CREATE") {
    if (el.beforeSave) {
      fs.writeFileSync("./beforeSave.js", el.beforeSave);
      const beforeSave = require("./beforeSave.js");
      app.post("/beforeSave", (req, res) => {
        const output = beforeSave(req.body);
        res.setHeader("Content-Type", "application/json");
        res.end(JSON.stringify(output));
      });
    }

    if (el.afterSave) {
      fs.writeFileSync("./afterSave.js", el.afterSave);
      const afterSave = require("./afterSave.js");
      app.post("/afterSave", (req, res) => {
        const output = afterSave(req.body);
        res.setHeader("Content-Type", "application/json");
        res.end(JSON.stringify(output));
      });
    }
  }
});

app.listen(port, () => console.log(`Listening on port ${port}`));
