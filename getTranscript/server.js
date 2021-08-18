const express = require("express");
const { gT } = require("./getTranscript");

const app = express();
const port = process.env.port || 5000;

//enabling body parser middleware for express
app.use(express.json());

//setting up the getTranscript handler
app.use("/", async (req, res) => {
  let youtubeURl = req.body.url;
  //to get the transcript
  console.log(`received request of url ${youtubeURl}`)
  let { result, err } = await gT(youtubeURl);
  if (err != undefined) {
    //incase of an error
    return res.status(500).json({ err: err.toString() });
  }
  res.status(200).json({ err: "", result: result });
});

//to start the server
app.listen(port, () => {
  console.log(`Transcript extractor listening at http://localhost:${port}`);
});
