'use strict';

const subscribeHook = async (z, bundle) => {
  const data = {
    url: bundle.targetUrl,
    events: ['deal.created'],
  };
  const response = await z.request({
    url: `${bundle.authData.baseUrl}/api/v1/webhooks`,
    method: 'POST',
    body: data,
  });
  return response.data;
};

const unsubscribeHook = async (z, bundle) => {
  const hookId = bundle.subscribeData.id;
  return z.request({
    url: `${bundle.authData.baseUrl}/api/v1/webhooks/${hookId}`,
    method: 'DELETE',
  });
};

const getFallbackPayload = async (z, bundle) => {
  const response = await z.request({
    url: `${bundle.authData.baseUrl}/api/v1/deals`,
    params: { limit: 1 },
  });
  const data = response.data.data || response.data || [];
  return data.slice(0, 1);
};

module.exports = {
  key: 'new_deal',
  noun: 'Deal',

  display: {
    label: 'New Deal',
    description: 'Triggers when a new deal is created in PraestOS.',
  },

  operation: {
    type: 'hook',
    performSubscribe: subscribeHook,
    performUnsubscribe: unsubscribeHook,
    perform: (z, bundle) => {
      const payload = bundle.cleanedRequest;
      return [payload];
    },
    performList: getFallbackPayload,

    sample: {
      id: 'a1b2c3d4-0000-0000-0000-000000000002',
      org_id: 'a1b2c3d4-0000-0000-0000-000000000000',
      title: 'Acme Corp — Enterprise Plan',
      value_cents: 1200000,
      currency: 'USD',
      stage: 'qualified',
      probability: 50,
      owner_id: 'a1b2c3d4-0000-0000-0000-000000000003',
      pipeline_id: 'a1b2c3d4-0000-0000-0000-000000000004',
      created_at: '2026-03-01T10:00:00Z',
      updated_at: '2026-03-01T10:00:00Z',
    },

    outputFields: [
      { key: 'id', label: 'Deal ID' },
      { key: 'title', label: 'Title' },
      { key: 'value_cents', label: 'Value (cents)' },
      { key: 'currency', label: 'Currency' },
      { key: 'stage', label: 'Stage' },
      { key: 'probability', label: 'Probability (%)' },
      { key: 'owner_id', label: 'Owner ID' },
      { key: 'pipeline_id', label: 'Pipeline ID' },
      { key: 'created_at', label: 'Created At' },
    ],
  },
};
