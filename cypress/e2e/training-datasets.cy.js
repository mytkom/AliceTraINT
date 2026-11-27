/// <reference types="cypress" />

describe('Training Dataset Management', () => {
    beforeEach(() => {
        cy.visit('/login');
    });

    const generateUniqueName = (prefix) => {
        const uuid = Cypress._.random(0, 1e6);
        return `${prefix}_${uuid}`;
    };

    const selectFiles = () => {
        cy.contains('#file-tree button', 'Small fixture run').click();
        cy.get('#tree-current-path').should('have.text', '/alice/sim/2024/LHC24f3/0/523397');
        cy.get('#switch-to-find-button').click();

        const rows = '#find-panel:visible [id="file-list"] li';
        cy.get(rows).should('have.length', 8);
        for (let i = 0; i < 5; i++) {
            cy.get(rows).first().click();
        }
        cy.get('#selected-files li').should('have.length', 5);
    };

    it('creates and removes a dataset', () => {
        cy.get('main a')
            .contains('Create Training Dataset')
            .click();

        const testName = generateUniqueName('LHC24f3');
        cy.get('#train-dataset-name').type(testName);

        selectFiles();

        cy.get('#submit-dataset-form button[type="submit"]')
            .contains('Submit')
            .click();

        const datasetEntry = cy.contains('#training-datasets-listing a', testName).closest('#training-datasets-listing > div')
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
        cy.get('#training-datasets-listing > div > a').first().invoke('text').then(name => {
            let alreadyExisting = name
            cy.wrap(alreadyExisting.trim()).as('alreadyExisting')
        })
        cy.get('main a')
            .contains('Create Training Dataset')
            .click();

        cy.get('@alreadyExisting').then(alreadyExisting => {
            cy.get('#train-dataset-name').type(alreadyExisting);
        })

        selectFiles();

        cy.get('#submit-dataset-form button[type="submit"]')
            .contains('Submit')
            .click();

        cy.get('#errors').should('contain.text', 'Name must be unique')
    });
});
