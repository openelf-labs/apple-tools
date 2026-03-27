ObjC.import("Foundation");

var params = JSON.parse($.NSProcessInfo.processInfo.environment.objectForKey("APPLE_TOOLS_PARAMS").js);
var targetName = params.name;

var app = Application("Contacts");
var people = app.people.whose({name: targetName});

if (people.length === 0) {
	JSON.stringify(null);
} else {
	var person = people[0];

	var phones = [];
	try {
		var pphones = person.phones();
		for (var i = 0; i < pphones.length; i++) {
			phones.push({label: pphones[i].label() || "", value: pphones[i].value()});
		}
	} catch (e) {}

	var emails = [];
	try {
		var pemails = person.emails();
		for (var i = 0; i < pemails.length; i++) {
			emails.push({label: pemails[i].label() || "", value: pemails[i].value()});
		}
	} catch (e) {}

	var addresses = [];
	try {
		var paddrs = person.addresses();
		for (var i = 0; i < paddrs.length; i++) {
			addresses.push({
				label: paddrs[i].label() || "",
				street: paddrs[i].street() || "",
				city: paddrs[i].city() || "",
				state: paddrs[i].state() || "",
				zip: paddrs[i].zip() || "",
				country: paddrs[i].country() || ""
			});
		}
	} catch (e) {}

	var org = "";
	try { org = person.organization() || ""; } catch (e) {}

	var title = "";
	try { title = person.jobTitle() || ""; } catch (e) {}

	var birthday = "";
	try {
		var bd = person.birthDate();
		if (bd) {
			birthday = bd.toISOString().split("T")[0];
		}
	} catch (e) {}

	var note = "";
	try { note = person.note() || ""; } catch (e) {}

	JSON.stringify({
		name: person.name(),
		phones: phones,
		emails: emails,
		addresses: addresses,
		organization: org,
		jobTitle: title,
		birthday: birthday,
		note: note
	});
}
