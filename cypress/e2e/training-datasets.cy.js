/// <reference types="cypress" />

describe('Training Dataset Management', () => {
    const baseUrl = 'http://localhost:8088/';

    beforeEach(() => {
        cy.setCookie('gosessionid', Cypress.env('GO_SESSION_ID'));
        cy.visit(baseUrl);
    });

    const generateUniqueName = (prefix) => {
        const uuid = Cypress._.random(0, 1e6);
        return `${prefix}_${uuid}`;
    };

    const selectFiles = (selectors) => {
        selectors.forEach(selector => {
            cy.get(selector).click();
        });
    };

    it('creates a dataset and remove it', () => {
        cy.get('main a')
            .contains('Create Training Dataset')
            .click();

        const testName = generateUniqueName('LHC24f3');
        cy.get('#train-dataset-name').type(testName);

        cy.get('#find-aods-form input').type('/alice/sim/2024/LHC24f3/0/523397');
        cy.get('#find-aods-form button').click();

        cy.wait(2000);
        cy.get('#file-list li').should('have.length', 22);

        selectFiles([
            '#file-list li:first-child',
            '#file-list li:first-child',
            '#file-list li:first-child',
            '#file-list li:first-child',
            '#file-list li:last-child'
        ]);

        cy.get('#submit-dataset-form button[type="submit"]')
            .contains('Submit')
            .click();

        cy.get('#training-datasets-listing button')
            .contains('Remove')
            .click();
    });

    it('disallows submission without a name', () => {
        cy.get('main a')
            .contains('Create Training Dataset')
            .click();

        cy.get('#submit-dataset-form button[type="submit"]')
            .contains('Submit')
            .click();

        cy.get('#train-dataset-name').should(($input) => {
            expect($input[0].checkValidity()).to.be.false;
            expect($input[0].validationMessage).to.contain('Please fill out this field.');
        });
    });
});
