/// <reference types="cypress" />

describe('Training Dataset Management', () => {
    const baseUrl = 'http://localhost:8088/login';

    beforeEach(() => {
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

        cy.get('#file-list ul').children().its('length').should('be.gt', 0);

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

        let datasetEntry = cy.get('#training-datasets-listing').children('div', testName).first()
        datasetEntry.should('exist')
        datasetEntry.within(() => cy.get('button').contains('Remove').click())

        cy.contains('#training-datasets-listing div', testName).should('not.exist')
    });

    it('userScoped', () => {
        cy.get('#training-datasets-listing').children().should('have.length', 2)
        cy.get('input[name="userScoped"]').click();
        cy.get('#training-datasets-listing').children().should('have.length', 1)
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

    it('validation error for duplicated name', () => {
        cy.get('#training-datasets-listing div div a div').first().invoke('text').then(name => {
            let alreadyExisting = name
            cy.wrap(alreadyExisting).as('alreadyExisting')
        })
        cy.get('main a')
            .contains('Create Training Dataset')
            .click();

        cy.get('@alreadyExisting').then(alreadyExisting => {
            cy.get('#train-dataset-name').type(alreadyExisting);
        })

        cy.get('#find-aods-form input').type('/alice/sim/2024/LHC24f3/0/523397');
        cy.get('#find-aods-form button').click();

        cy.get('#file-list ul').children().its('length').should('be.gt', 0);

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

        cy.get('#errors').invoke('text').should('eq', 'Name must be unique\n')
    });
});
