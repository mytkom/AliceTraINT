/// <reference types="cypress" />

describe('Training Tasks Management', () => {
    beforeEach(() => {
        cy.visit('/login');
        cy.visit('/training-tasks');
    });

    const generateUniqueName = (prefix) => {
        const uuid = Cypress._.random(0, 1e6);
        return `${prefix}_${uuid}`;
    };

    it('creates a task', () => {
        cy.get('main a')
            .contains('Create Training Task')
            .click();

        const testName = generateUniqueName('tm');
        cy.get('input[name="name"]').type(testName);
        cy.get('select[name="trainingDatasetId"]').select('1');
        cy.get('button[type="submit"]').click();

        let tmObject = cy.contains('tr', testName)
        tmObject.should('exist')
    });

    it('userScoped', () => {
        cy.contains('td', 'Niels Bohr').should('exist')
        cy.get('input[name="userScoped"]').click();
        cy.contains('td', 'Niels Bohr').should('not.exist')
    });

    it('disallows submission without a name', () => {
        cy.get('main a')
            .contains('Create Training Task')
            .click();

        cy.get('button[type="submit"]')
            .click();

        cy.get('input[name="name"]').should(($input) => {
            expect($input[0].checkValidity()).to.be.false;
            expect($input[0].validationMessage).to.contain('Please fill out this field.');
        });
    });

    it('validation error for duplicated name', () => {
        cy.get('tbody tr').first().children().first().invoke('text').then(name => {
            let alreadyExisting = name
            cy.wrap(alreadyExisting.trim()).as('alreadyExisting')
        })

        cy.get('main a')
            .contains('Create Training Task')
            .click();

        cy.get('@alreadyExisting').then(alreadyExisting => {
            cy.get('input[name="name"]').type(alreadyExisting);
        })
        cy.get('select[name="trainingDatasetId"]').select('1');
        cy.get('button[type="submit"]').click();

        cy.get('#errors').should('contain.text', 'Name must be unique')
    });

    it('opens task details from the linked name and renders read-only statuses', () => {
        cy.get('.app-nav a[aria-current="page"]').should('have.text', 'Tasks');
        cy.contains('tr', 'LHC24b1b Task 2').within(() => {
            cy.get('.task-status').should('match', 'span').and('have.text', 'Completed');
            cy.get('.task-status').should('not.have.attr', 'tabindex');
            cy.contains('a', 'LHC24b1b Task 2').click();
        });
        cy.contains('h2', 'LHC24b1b Task 2').should('be.visible');
        cy.get('.app-nav a[aria-current="location"]').should('have.text', 'Tasks');
    });

    it('opens and closes mobile navigation and follows a section link', () => {
        cy.viewport(390, 844);
        cy.get('#app-navigation').should('not.be.visible');
        cy.get('#menu-button').should('have.attr', 'aria-expanded', 'false').click();
        cy.get('#menu-button').should('have.attr', 'aria-expanded', 'true');
        cy.focused().should('have.text', 'Tasks').trigger('keydown', { key: 'Escape' });
        cy.get('#app-navigation').should('not.be.visible');
        cy.get('#menu-button').should('be.focused').click();
        cy.contains('#app-navigation a', 'Datasets').click();
        cy.location('pathname').should('eq', '/training-datasets');
        cy.get('#menu-button').click();
        cy.get('.app-nav a[aria-current="page"]').should('have.text', 'Datasets');
    });

    it('shows loading and failure feedback and recovers on the next request', () => {
        cy.intercept({ pathname: '/training-tasks/list', times: 1 }, {
            statusCode: 500,
            delay: 700,
        }).as('failedList');
        cy.reload();
        cy.contains('[role="status"]', 'Loading training tasks').should('be.visible');
        cy.wait('@failedList');
        cy.contains('[role="alert"]', 'Training tasks could not be loaded.').should('be.visible');
        cy.contains('label', 'Show only mine').click();
        cy.contains('a', 'LHC24b1b Task 2').should('be.visible');
        cy.get('#task-list-error').should('not.be.visible');
    });

    for (const [size, width, height] of [['desktop', 1440, 900], ['mobile', 390, 844]]) {
        it(`renders both themes at ${size} size and remembers the selection`, () => {
            cy.viewport(width, height);
            cy.window().then(win => win.localStorage.setItem('alicetraint-theme', 'light'));
            cy.reload();
            cy.contains('a', 'LHC24b1b Task 2').should('be.visible');
            cy.get('.app-shell').should('have.css', 'background-color', 'rgb(229, 237, 248)');
            cy.get('#theme-button').should('have.attr', 'aria-pressed', 'false').click();
            cy.get('#theme-button').should('have.attr', 'aria-pressed', 'true');
            cy.get('.app-shell').should('have.css', 'background-color', 'rgb(16, 28, 43)');
            cy.reload();
            cy.get('#theme-button').should('have.attr', 'aria-pressed', 'true');
            cy.get('.task-status[data-status="Completed"]').first()
                .should('have.css', 'background-color', 'rgb(32, 61, 57)');
            cy.document().then(doc => {
                expect(doc.documentElement.scrollWidth).to.equal(doc.documentElement.clientWidth);
            });
        });
    }

});
