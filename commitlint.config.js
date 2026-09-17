module.exports = {
  extends: ['@commitlint/config-conventional'],
  rules: {
    'type-enum': [
      2,
      'always',
      [
        'feat',
        'fix',
        'chore',
        'docs',
        'refactor',
        'test',
        'style',
        'perf',
        'ci',
        'build',
        'revert'
      ]
    ],
    'subject-empty': [2, 'never'],
    'type-empty': [2, 'never']
  }
};

