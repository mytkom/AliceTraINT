/// <reference types="cypress" />

describe('Training Machine Management', () => {
    const baseUrl = 'http://localhost:8088';

    beforeEach(() => {
        cy.visit(`${baseUrl}/login`);
        cy.visit(`${baseUrl}/training-machines`);
    });

    const generateUniqueName = (prefix) => {
        const uuid = Cypress._.random(0, 1e6);
        return `${prefix}_${uuid}`;
    };

    it('creates a machine and remove it', () => {
        cy.get('main a')
            .contains('Register Training Machine')
            .click();

        const testName = generateUniqueName('tm');
        cy.get('input[name="name"]').type(testName);
        cy.get('button').click();
        cy.get('button').click();

        let tmObject = cy.contains('tr', testName) 
        tmObject.should('exist')
        tmObject.within(() => cy.get('button').contains('Remove').click())
        cy.contains('tr', testName).should('not.exist')
    });

    it('userScoped', () => {
        cy.contains('td', 'Niels Bohr').should('exist')
        cy.get('input[name="userScoped"]').click();
        cy.contains('td', 'Niels Bohr').should('not.exist')
    });

    it('disallows submission without a name', () => {
        cy.get('main a')
            .contains('Register Training Machine')
            .click();

        cy.get('button[type="submit"]')
            .click();

        cy.get('input').should(($input) => {
            expect($input[0].checkValidity()).to.be.false;
            expect($input[0].validationMessage).to.contain('Please fill out this field.');
        });
    });
});