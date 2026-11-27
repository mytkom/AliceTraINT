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
});
