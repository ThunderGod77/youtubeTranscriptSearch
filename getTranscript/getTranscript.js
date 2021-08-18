//to use a headless chrome instance
const puppeteer = require("puppeteer");

//function to get transcription of a youtube video
exports.gT = async function getTranscript(url) {
  let browser;
  let page;
  try {
    //to launch the browser
    browser = await puppeteer.launch();
  } catch (err) {
    const e = new Error("Something went wrong!");
    //to close the browser instance
    console.log(err);
    return { err: e, result: undefined };
  }
  try {
    page = await browser.newPage();
  } catch (err) {
    const e = new Error("Something went wrong!");
    await browser.close();
    return { err: e, result: undefined };
  }
  try {
    //to visit the webpage corresponding to the youtube URL
    await page.goto(url);
  } catch (err) {
    //incase of an invalid url
    const e = new Error("Cannot find the url!");
    await browser.close();
    return { err: e, result: undefined };
  }
  try {
    //to check if the webpage contains a button only present in the page containing video
    await page.waitForSelector(
      "ytd-menu-renderer.ytd-video-primary-info-renderer > yt-icon-button:nth-child(2) > button:nth-child(1)"
    );
  } catch (err) {
    //incase the youtube url does not contain a video
    const e = new Error("Incorrect youtube video url!");
    await browser.close();
    return { err: e, result: undefined };
  }
  //to click on the three dot button below a video left of save button
  await page.click(
    "ytd-menu-renderer.ytd-video-primary-info-renderer > yt-icon-button:nth-child(2) > button:nth-child(1)"
  );
  try {
    //to check if the video has a trancript option or not
    await page.waitForSelector(
      "ytd-menu-service-item-renderer.style-scope:nth-child(2)"
    );
  } catch (err) {
    //return an error if the video does not have a transcript
    const e = new Error("The video does not have a transcript!");
    await browser.close();
    return { err: e, result: undefined };
  }
  await page.click("ytd-menu-service-item-renderer.style-scope:nth-child(2)");
  //to wait for the transcript menu to load
  await page.waitForSelector("div.cue-group:nth-child(1)");
  let result;
  try {
    result = await page.evaluate(async () => {
      //to store title of the video
      let vidTitle = await document.querySelector(
        "yt-formatted-string.ytd-video-primary-info-renderer:nth-child(1)"
      ).textContent;
      //to store the channel name
      let vidChannel = await document.querySelector(
        "ytd-video-owner-renderer.ytd-video-secondary-info-renderer > div:nth-child(2) > ytd-channel-name:nth-child(1) > div:nth-child(1) > div:nth-child(1) > yt-formatted-string:nth-child(1) > a:nth-child(1)"
      ).textContent;
      //to store timestamps of the transcript
      let timeStamps = [];
      //to store the captions
      let caption = [];
      //to get all the div elements containing the timestamp
      let allTime = await document.querySelectorAll(".cue-group-start-offset");
      allTime = Array.from(allTime);
      //to extract the timestamp from each div tag
      allTime.forEach((e) => {
        //adding the timestamps
        timeStamps.push(e.textContent);
      });

      //to get all the div elements containing captions
      let allCaption = await document.querySelectorAll(
        ".ytd-transcript-body-renderer"
      );
      allCaption = Array.from(allCaption);
      allCaption.forEach((e) => {
        //filtering elemnts do not contain the captions
        if (e.getAttribute("role") == "button") {
          //adding the captions
          caption.push(e.textContent);
        }
      });

      return {
        timeStamps: timeStamps,
        caption: caption,
        vidTitle: vidTitle,
        vidChannel: vidChannel,
      };
    });
  } catch (error) {
    const e = new Error("Internal error!");
    await browser.close();
    return { err: e, result: undefined };
  }
  //to close the browser instance
  await browser.close();

  return { err: undefined, result: result };
};
// to test the getTranscript function
async function lol() {
  let { err } = await getTranscript(
    "https://www.youtube.com/watch?v=yj46BWpxFcA"
  );
}
