// jshint esversion:6
const Mailgun = require('mailgun-js');
const humanizeDuration = require('humanize-duration');
const config = require('./config.json');

const mailgun = Mailgun({
  apiKey: config.MAILGUN_API_KEY,
  domain: config.MAILGUN_DOMAIN,
});

// subscribeMailgun is the main function called by Cloud Functions.
module.exports.subscribeMailgun = (pubSubEvent, context) => {
  const build = eventToBuild(pubSubEvent.data);

  // Skip if the current status is not in the status list.
  const status = ['SUCCESS', 'FAILURE', 'INTERNAL_ERROR', 'TIMEOUT'];
  if (status.indexOf(build.status) === -1) {
    return;
  }

  if (build.steps[0].name == "gcr.io/cloud-builders/gcs-fetcher"){
    return;
  }

  // Send email.
  const message = createEmail(build);
  mailgun.messages().send(message, (error, body) => console.log(body.message));
};

// eventToBuild transforms pubsub event message to a build object.
const eventToBuild = (data) => {
  return JSON.parse(Buffer.from(data, 'base64').toString());
};

// createEmail creates an email message from a build object.
const createEmail = (build) => {

    
  let color = '#00FF00';
  if (build.status != "SUCCESS"){
    color = '#FF0000';
  }

  let weight = '400';
  if (build.status != "SUCCESS"){
    weight = '900';
  }

  let buildlink = `<a href="${build.logUrl}">${build.id}</a>`;
  
  const duration = humanizeDuration(new Date(build.finishTime) - new Date(build.startTime));
  
  
  let msgHtml = `<p><a href="${build.logUrl}">Latest build</a> finished with status <span style="color: ${color}; font-weight: ${weight}">${build.status} </span>, in ${duration}.</p>
                <p>View it here: <a href="https://${process.env.GCLOUD_PROJECT}.appspot.com">https://${process.env.GCLOUD_PROJECT}.appspot.com</a></p>
  
  
                `;
  
  console.log(build);
  
  const message = {
    from: config.MAILGUN_FROM,
    to: config.MAILGUN_TO,
    subject: `${build.status}: Project: ${build.projectId} Build finished`,
    text: "Build complete",
    html: msgHtml
  };
  return message;
};