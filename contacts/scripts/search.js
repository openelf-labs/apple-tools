ObjC.import("Foundation");

var params = JSON.parse($.NSProcessInfo.processInfo.environment.objectForKey("APPLE_TOOLS_PARAMS").js);
var query = params.query;
var limit = params.limit || 25;

var app = Application("Contacts");
var people = app.people.whose({name: {_contains: query}});
var count = Math.min(people.length, limit);
var results = [];

for (var i = 0; i < count; i++) {
	var person = people[i];
	var phones = [];
	var emails = [];

	try {
		var pphones = person.phones();
		for (var j = 0; j < pphones.length; j++) {
			phones.push(pphones[j].value());
		}
	} catch (e) {}

	try {
		var pemails = person.emails();
		for (var j = 0; j < pemails.length; j++) {
			emails.push(pemails[j].value());
		}
	} catch (e) {}

	var org = "";
	try { org = person.organization() || ""; } catch (e) {}

	var title = "";
	try { title = person.jobTitle() || ""; } catch (e) {}

	results.push({
		name: person.name(),
		phones: phones,
		emails: emails,
		organization: org,
		jobTitle: title
	});
}

JSON.stringify(results);
