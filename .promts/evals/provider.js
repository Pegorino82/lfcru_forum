/**
 * Fixture provider — возвращает заранее записанные ответы вместо API-вызовов.
 * config.fixture: 'broken' | 'fixed'
 */
const fs = require('fs');
const path = require('path');

class FixtureProvider {
  constructor(options) {
    this.fixtureType = options?.config?.fixture || 'broken';
  }

  id() {
    return `fixture:${this.fixtureType}`;
  }

  async callApi(prompt, context) {
    const cardContext = context.vars?.card_context || '';
    const isThinCard = /Description:\s*""/.test(cardContext);
    const suffix = isThinCard ? '-thin' : '';
    const fixtureFile = path.join(__dirname, 'fixtures', `${this.fixtureType}${suffix}.md`);
    let content = fs.readFileSync(fixtureFile, 'utf-8');

    // Подставляем конкретный FT-номер из card_context
    const ftMatch = cardContext.match(/FT-(\d+)/);
    if (ftMatch) {
      content = content.replace(/FT-XXX/g, `FT-${ftMatch[1]}`);
    }

    return { output: content };
  }
}

module.exports = FixtureProvider;
