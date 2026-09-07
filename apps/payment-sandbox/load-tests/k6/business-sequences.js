import http from 'k6/http';
import { check, sleep } from 'k6';

const baseUrl = normalizeBaseUrl(__ENV.BASE_URL || 'http://localhost:8080/v1');
const scenario = __ENV.SCENARIO || 'approved_immediate';
const captureMethod = __ENV.CAPTURE_METHOD || 'manual';
const amount = Number.parseInt(__ENV.AMOUNT || '1000', 10);
const currency = __ENV.CURRENCY || 'usd';
const merchantId = __ENV.MERCHANT_ID || 'merchant_123';
const customerId = __ENV.CUSTOMER_ID || 'cus_123';
const paymentMethodToken = __ENV.PAYMENT_METHOD_TOKEN || 'pm_card_visa';

export const options = {
	vus: Number.parseInt(__ENV.VUS || '1', 10),
	duration: __ENV.DURATION || '1m',
	thresholds: {
		http_req_failed: ['rate<0.01'],
		http_req_duration: ['p(95)<500'],
	},
};

export default function () {
	const intent = createPaymentIntent();
	confirmPaymentIntent(intent.id);

	if (captureMethod === 'automatic') {
		capturePaymentIntent(intent.id);
	}

	sleep(Number.parseFloat(__ENV.SLEEP || '1'));
}

function createPaymentIntent() {
	const payload = JSON.stringify({
		amount,
		currency,
		merchant_id: merchantId,
		customer_id: customerId,
		capture_method: captureMethod,
		idempotency_key: `create-${scenario}-${__VU}-${__ITER}`,
	});

	const res = http.post(`${baseUrl}/payment_intents`, payload, jsonHeaders());
	check(res, {
		'create payment intent status is 201': (response) => response.status === 201,
	});

	return extractEntity(res);
}

function confirmPaymentIntent(id) {
	const payload = JSON.stringify({
		payment_method_token: paymentMethodToken,
		idempotency_key: `confirm-${scenario}-${__VU}-${__ITER}`,
	});

	const res = http.post(`${baseUrl}/payment_intents/${id}/confirm`, payload, {
		headers: {
			...jsonHeaders().headers,
			'X-Sandbox-Scenario': scenario,
		},
	});
	check(res, {
		'confirm payment intent status is 200': (response) => response.status === 200,
	});

	return extractEntity(res);
}

function capturePaymentIntent(id) {
	const res = http.post(`${baseUrl}/payment_intents/${id}/capture`, null, jsonHeaders());
	check(res, {
		'capture payment intent status is 200': (response) => response.status === 200,
	});

	return extractEntity(res);
}

function extractEntity(response) {
	const body = response.json();
	if (!body) {
		return {};
	}

	return body.data || body.payment_intent || body.paymentIntent || body.intent || body;
}

function jsonHeaders() {
	return {
		headers: {
			'Content-Type': 'application/json',
		},
	};
}

function normalizeBaseUrl(value) {
	return value.replace(/\/+$/, '');
}
