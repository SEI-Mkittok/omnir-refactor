'use strict';

const { version } = require('zapier-platform-core');

const newTicket = require('./triggers/new_ticket');
const newDeal = require('./triggers/new_deal');
const newContact = require('./triggers/new_contact');
const ticketStatusChanged = require('./triggers/ticket_status_changed');

const createTicket = require('./actions/create_ticket');
const createContact = require('./actions/create_contact');
const updateDealStage = require('./actions/update_deal_stage');

const App = {
  version: require('./package.json').version,
  platformVersion: version,

  authentication: {
    type: 'custom',
    fields: [
      {
        key: 'apiKey',
        label: 'API Key',
        required: true,
        type: 'password',
        helpText:
          'Your PraestOS API key. Generate one at **Settings → API Keys** in the PraestOS dashboard.',
      },
      {
        key: 'baseUrl',
        label: 'Base URL',
        required: true,
        type: 'string',
        default: 'https://app.praestos.io',
        helpText:
          'The base URL of your PraestOS instance (no trailing slash). Self-hosted customers should enter their own domain.',
      },
    ],

    test: async (z, bundle) => {
      const response = await z.request({
        url: `${bundle.authData.baseUrl}/api/v1/users`,
        params: { limit: 1 },
      });
      return response.data;
    },

    connectionLabel: (z, bundle) => bundle.authData.baseUrl,
  },

  beforeRequest: [
    (request, z, bundle) => {
      request.headers['Authorization'] = `Bearer ${bundle.authData.apiKey}`;
      request.headers['Content-Type'] = 'application/json';
      return request;
    },
  ],

  triggers: {
    [newTicket.key]: newTicket,
    [newDeal.key]: newDeal,
    [newContact.key]: newContact,
    [ticketStatusChanged.key]: ticketStatusChanged,
  },

  creates: {
    [createTicket.key]: createTicket,
    [createContact.key]: createContact,
    [updateDealStage.key]: updateDealStage,
  },
};

module.exports = App;
