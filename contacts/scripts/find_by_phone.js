ObjC.import("Foundation");

var params = JSON.parse($.NSProcessInfo.processInfo.environment.objectForKey("APPLE_TOOLS_PARAMS").js);
var phoneNumber = params.phoneNumber;

// Normalize: strip everything except digits and leading +
function normalizePhone(num) {
	var hasPlus = num.charAt(0) === "+";
	var digits = num.replace(/[^\d]/g, "");
	return hasPlus ? "+" + digits : digits;
}

var normalized = normalizePhone(phoneNumber);

// Generate match variants: original digits, without country code, with +1, etc.
function phoneVariants(norm) {
	var variants = [norm];
	var digits = norm.replace(/^\+/, "");

	if (digits !== norm) variants.push(digits);

	// If has country code (starts with 1 and 11 digits, or starts with other and 12+ digits)
	if (digits.length >= 11 && digits.charAt(0) === "1") {
		variants.push(digits.substring(1));
		variants.push("+" + digits);
	}
	// If 10 digits (US), try with +1 prefix
	if (digits.length === 10) {
		variants.push("1" + digits);
		variants.push("+1" + digits);
	}

	return variants;
}

var variants = phoneVariants(normalized);

var app = Application("Contacts");
var allPeople = app.people();
var found = null;

for (var i = 0; i < allPeople.length && !found; i++) {
	var person = allPeople[i];
	var pphones;
	try {
		pphones = person.phones();
	} catch (e) {
		continue;
	}

	for (var j = 0; j < pphones.length && !found; j++) {
		var pval = normalizePhone(pphones[j].value());
		for (var k = 0; k < variants.length; k++) {
			if (pval === variants[k] || pval.indexOf(variants[k]) !== -1 || variants[k].indexOf(pval) !== -1) {
				var phones = [];
				try {
					var pp = person.phones();
					for (var m = 0; m < pp.length; m++) phones.push(pp[m].value());
				} catch (e) {}

				var emails = [];
				try {
					var pe = person.emails();
					for (var m = 0; m < pe.length; m++) emails.push(pe[m].value());
				} catch (e) {}

				found = {
					name: person.name(),
					phones: phones,
					emails: emails
				};
				break;
			}
		}
	}
}

JSON.stringify(found);
