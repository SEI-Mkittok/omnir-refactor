'use strict';

const updateDealStage = async (z, bundle) => {
  const response = await z.request({
    url: `${bundle.authData.baseUrl}/api/v1/deals/${bundle.inputData.deal_id}`,
    method: 'PATCH',
    body: {
      stage: bundle.inputData.stage,
    },
  });
  return response.data;
};

module.exports = {
  key: 'update_deal_stage',
  noun: 'Deal',

  display: {
    label: 'Update Deal Stage',
    description: 'Updates the pipeline stage of an existing deal in PraestOS.',
  },

  operation: {
    perform: updateDealStage,

    inputFields: [
      {
        key: 'deal_id',
        label: 'Deal ID',
        type: 'string',
        required: true,
        helpText: 'The UUID of the deal to update.',
      },
      {
        key: 'stage',
        label: 'Stage',
        type: 'string',
        required: true,
        choices: ['lead', 'qualified', 'proposal', 'negotiation', 'closed_won', 'closed_lost'],
        helpText: 'The new pipeline stage for the deal.',
      },
    ],

    sample: {
      id: 'a1b2c3d4-0000-0000-0000-000000000002',
      org_id: 'a1b2c3d4-0000-0000-0000-000000000000',
      title: 'Acme Corp — Enterprise Plan',
      value_cents: 1200000,
      currency: 'USD',
      stage: 'closed_won',
      probability: 100,
      created_at: '2026-03-01T10:00:00Z',
      updated_at: '2026-03-01T12:00:00Z',
    },
  },
};
