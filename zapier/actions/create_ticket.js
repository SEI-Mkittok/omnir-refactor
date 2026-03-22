'use strict';

const createTicket = async (z, bundle) => {
  const response = await z.request({
    url: `${bundle.authData.baseUrl}/api/v1/tickets`,
    method: 'POST',
    body: {
      subject: bundle.inputData.subject,
      description: bundle.inputData.description || undefined,
      status: bundle.inputData.status || 'open',
      priority: bundle.inputData.priority || 'medium',
      contact_id: bundle.inputData.contact_id || undefined,
      assignee_id: bundle.inputData.assignee_id || undefined,
      tags: bundle.inputData.tags ? bundle.inputData.tags.split(',').map((t) => t.trim()) : [],
    },
  });
  return response.data;
};

module.exports = {
  key: 'create_ticket',
  noun: 'Ticket',

  display: {
    label: 'Create Ticket',
    description: 'Creates a new support ticket in PraestOS.',
  },

  operation: {
    perform: createTicket,

    inputFields: [
      {
        key: 'subject',
        label: 'Subject',
        type: 'string',
        required: true,
        helpText: 'The subject line of the ticket.',
      },
      {
        key: 'description',
        label: 'Description',
        type: 'text',
        required: false,
        helpText: 'Detailed description of the issue.',
      },
      {
        key: 'status',
        label: 'Status',
        type: 'string',
        required: false,
        default: 'open',
        choices: ['open', 'in_progress', 'pending', 'resolved', 'closed'],
        helpText: 'Initial ticket status.',
      },
      {
        key: 'priority',
        label: 'Priority',
        type: 'string',
        required: false,
        default: 'medium',
        choices: ['low', 'medium', 'high', 'critical'],
        helpText: 'Ticket priority.',
      },
      {
        key: 'contact_id',
        label: 'Contact ID',
        type: 'string',
        required: false,
        helpText: 'UUID of the contact to associate with this ticket.',
      },
      {
        key: 'assignee_id',
        label: 'Assignee ID',
        type: 'string',
        required: false,
        helpText: 'UUID of the agent to assign this ticket to.',
      },
      {
        key: 'tags',
        label: 'Tags',
        type: 'string',
        required: false,
        helpText: 'Comma-separated list of tags.',
      },
    ],

    sample: {
      id: 'a1b2c3d4-0000-0000-0000-000000000001',
      org_id: 'a1b2c3d4-0000-0000-0000-000000000000',
      subject: 'Cannot log in to my account',
      status: 'open',
      priority: 'medium',
      tags: [],
      created_at: '2026-03-01T10:00:00Z',
      updated_at: '2026-03-01T10:00:00Z',
    },
  },
};
