describe('Public landing and documentation', () => {
    const setLightTheme = (win) => win.localStorage.setItem('alicetraint-theme', 'light');

    it('signs in through the fixture SSO and redirects signed-in visitors from home', () => {
        cy.visit('/', { onBeforeLoad: setLightTheme });
        cy.contains('h1', 'Train and review PID models').should('be.visible');
        cy.contains('Example training task').should('not.exist');
        cy.contains('a', 'Sign in with CERN SSO').should('have.attr', 'href', '/login').click();
        cy.location('pathname').should('eq', '/training-datasets');
        cy.visit('/');
        cy.location('pathname').should('eq', '/training-datasets');
    });

    it('opens public docs and follows section and page links with the mobile navigation', () => {
        cy.viewport(390, 700);
        cy.visit('/', { onBeforeLoad: setLightTheme });
        cy.get('#menu-button').click().should('have.attr', 'aria-expanded', 'true');
        cy.get('#app-navigation').contains('a', 'Docs').should('be.visible');
        cy.get('#app-navigation').contains('a', 'Sign in').should('have.attr', 'href', '/login');
        cy.get('#app-navigation a').first().trigger('keydown', { key: 'Escape' });
        cy.get('#menu-button').should('be.focused').and('have.attr', 'aria-expanded', 'false');
        cy.get('#theme-button').click();
        cy.contains('a', 'Read the docs').click();
        cy.location('pathname').should('eq', '/docs/getting-started');
        cy.get('html').should('have.attr', 'data-theme', 'dark');
        cy.contains('h1', 'Getting started').should('be.visible');
        cy.get('#docs-navigation').should('not.be.visible');
        cy.contains('button', 'Browse documentation').click().should('have.attr', 'aria-expanded', 'true');
        cy.get('#docs-navigation [aria-current="page"]').should('contain.text', 'Getting started');
        cy.get('#docs-navigation').contains('a', 'How it works').click();
        cy.location('hash').should('eq', '#sec-how-it-works');
        cy.get('article pre.mermaid > svg').should('be.visible');
        cy.get('#docs-navigation').contains('a', 'Local development').click();
        cy.location('pathname').should('eq', '/docs/local-development');
        cy.get('article pre .docs-copy-button').first().should('have.text', 'Copy');
        cy.document().should((doc) => {
            expect(doc.documentElement.scrollWidth).to.equal(doc.documentElement.clientWidth);
        });
    });

    it('keeps the visible Mermaid diagram in place until the new theme is ready', () => {
        cy.visit('/docs/getting-started', { onBeforeLoad: setLightTheme });
        cy.contains('article code', 'wut-alice-pidml').should('have.css', 'background-color', 'rgb(186, 230, 253)');
        cy.get('article pre.mermaid > svg').should('be.visible').then(($diagram) => {
            const original = $diagram[0];
            let release;
            cy.window().then((win) => {
                const render = win.mermaid.render;
                cy.stub(win.mermaid, 'render').callsFake((...args) =>
                    new Promise((resolve) => {
                        release = () => resolve(render.apply(win.mermaid, args));
                    })
                ).as('renderMermaid');
            });
            cy.get('#theme-button').click().should('have.attr', 'aria-pressed', 'true');
            cy.get('@renderMermaid').should('have.been.calledOnce');
            cy.contains('article code', 'wut-alice-pidml').should('have.css', 'background-color', 'rgb(7, 89, 133)');
            cy.get('article pre.mermaid > svg').should('be.visible').and(($current) => {
                expect($current[0]).to.equal(original);
            });
            cy.then(() => { release(); });
            cy.get('article pre.mermaid > svg').should('be.visible').and(($current) => {
                expect($current[0]).not.to.equal(original);
                expect($current.find('style').text()).not.to.equal(original.querySelector('style').textContent);
            });
        });
    });
});
