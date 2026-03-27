ObjC.import("Foundation");

var params = JSON.parse(
  $.NSProcessInfo.processInfo.environment
    .objectForKey("APPLE_TOOLS_PARAMS").js
);

var messages = Application("Messages");
var targetService = null;
var services = messages.services();

for (var i = 0; i < services.length; i++) {
  var svc = services[i];
  if (svc.serviceType() === "iMessage") {
    targetService = svc;
    break;
  }
}

if (!targetService) {
  JSON.stringify({ success: false, message: "iMessage service not found. Ensure you are signed in to iMessage." });
} else {
  var buddy = targetService.buddies.whose({ handle: params.phoneNumber })();
  var target;

  if (buddy.length > 0) {
    target = buddy[0];
    messages.send(params.message, { to: target });
    JSON.stringify({ success: true, message: "Message sent to " + params.phoneNumber });
  } else {
    // For new conversations, send directly without a buddy reference.
    // The Messages app will create the conversation automatically.
    messages.send(params.message, { to: targetService.buddies.whose({ handle: params.phoneNumber }) });
    JSON.stringify({ success: true, message: "Message sent to " + params.phoneNumber });
  }
}
