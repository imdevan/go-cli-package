import config from './config.mjs';

const apiReference = {
  label: 'API Reference',
  items: [
    { label: 'app', link: '/api/app' },
    { label: 'errors', link: '/api/errors' },
    {
      label: 'Adapters',
      items: [],
    },
  ],
};

const sidebar = [
  { label: 'go-cli-package', link: '/' },
  { label: 'Install', link: '/install' },
  { label: 'Commands', items: [
    { label: 'go-cli-package', link: '/commands/go-cli-package' },
    { label: 'completion', link: '/commands/completion' },

  ] },
  { label: 'Configuration', link: '/configuration' },
];

const isProduction = process.env.NODE_ENV === 'production';
if (!isProduction) {
  sidebar.push(apiReference);
}

export default sidebar;
