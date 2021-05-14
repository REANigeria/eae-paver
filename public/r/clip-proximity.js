import '../lib/selectlist.js';

import * as inputs from '../inputs.js';

import * as main from './main.js';

const header  = "Clip/Proximity";

const payload = {
	geographyid: null,
	datasetid: null,
	dataseturl: null,
	referenceurl: null,
	fields: null,
};

main.setup({ header, payload });

inputs.geographies({
	after: x => datasetid(x),
	payload
});

function datasetid(oldinput) {
	return inputs.datasetid({
		before: _ => oldinput.remove(),
		after: t => datasetinput(t),
		payload
	});
};

function datasetinput(oldinput) {
	return inputs.url({
		label: 'dataseturl',
		before: _ => oldinput.remove(),
		after: t => attrsinput(t),
		payload
	});
};

function attrsinput(oldinput) {
	return inputs.fields({
		label: 'fields',
		before: _ => oldinput.remove(),
		after: _ => main.submit('clip-proximity'),
		payload
	});
};