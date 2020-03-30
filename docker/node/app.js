const fs = require("fs");

const express = require("express");
const bodyParser = require("body-parser");

const app = express();
const port = 8080;

app.use(bodyParser.json());

const files = fs.readdirSync("./customLogic/");
files.filter(file =>
  file.endsWith(".js").forEach(file => {
    const customLogic = require(file);
    const fileNoExt = file.substring(0, file.length - 3);
    app.post("/" + fileNoExt, (req, res) => {
      const output = customLogic(req.body);
      res.setHeader("Content-Type", "application/json");
      res.end(JSON.stringify(output));
    });
  })
);

app.listen(port, () => console.log(`Listening on port ${port}`));
