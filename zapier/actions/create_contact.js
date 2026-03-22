'use strict';

const createContact = async (z, bundle) => {
  const response = await z.request({
    url: `${bundle.authData.baseUrl}/api/v1/contacts`,
    method: 'POST',
    body: {
      first_name: bundle.inputData.first_name,
      last_name: bundle.inputData.last_name,
      email: bundle.inputData.email || undefined,
      phone: bundle.inputData.phone || undefined,
      lead_source: bundle.inputData.lead_source || undefined,
      stage: bundle.inputData.stage || 'lead',
    },
  });
  return response.data;
};

module.exports = {
  key: 'create_contact',
  noun: 'Contact',

  display: {
    label: 'Create Contact',
    description: 'Creates a new contact in PraestOS.',
  },

  operation: {
    perform: createContact,

    inputFields: [
      {
        key: 'first_name',
        label: 'First Name',
        type: 'string',
        required: true,
      },
      {
        key: 'last_name',
        label: 'Last Name',
        type: 'string',
        required: true,
      },
      {
        key: 'email',
        label: 'Email',
        type: 'string',
        required: false,
      },
      {
        key: 'phone',
        label: 'Phone',
        type: 'string',
        required: false,
      },
      {
        key: 'lead_source',
        label: 'Lead Source',
        type: 'string',
        required: false,
        helpText: 'Where this contact came from (e.g. "Zapier", "Website", "Referral").',
      },
      {
        key: 'stage',
        label: 'Stage',
        type: 'string',
        required: false,
        default: 'lead',
        choices: ['lead', 'prospect', 'customer', 'churned'],
        helpText: 'Contact lifecycle stage.',
      },
    ],

    sample: {
      id: 'a1b2c3d4-0000-0000-0000-000000000005',
      org_id: 'a1b2c3d4-0000-0000-0000-000000000000',
      first_name: 'Jane',
      last_name: 'Doe',
      email: 'jane.doe@example.com',
      stage: 'lead',
      lead_score: 0,
      tags: [],
      email_opt_out: false,
      created_at: '2026-03-01T10:00:00Z',
      updated_at: '2026-03-01T10:00:00Z',
    },
  },
};
