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

    it('submits edited collapsed groups and restores them when reusing a configuration', () => {
        cy.contains('main a', 'Create Training Task').click();

        const taskName = generateUniqueName('configured');
        cy.get('#name').type(taskName);
        cy.get('#trainingDatasetId').select('1');
        cy.contains('summary', /^Events$/).click();
        cy.get('#enableSubsampleEventCount').check();
        cy.get('#subsampleEventCountInput').clear().type('12345');
        cy.contains('summary', /^Model$/).click();
        cy.get('#subsampleEventCountInput').should('be.visible');

        const architecture = '[name="configuration.model.architecture"]';
        const dropout = '[name="configuration.model.mlp.dropout"]';
        cy.get(architecture).select('mlp');
        cy.get('[name="configuration.model.attention.dropout"]').should('be.disabled');
        cy.get(dropout).clear().type('0.25');
        cy.contains('summary', /^Model$/).click();
        cy.contains('summary', /^Events$/).click();
        cy.contains('summary', /^Model$/).parent().should('have.prop', 'open', false);
        cy.contains('summary', /^Events$/).parent().should('have.prop', 'open', false);
        cy.get(dropout).should('be.enabled');

        cy.intercept('POST', '/training-tasks').as('createConfiguredTask');
        cy.contains('button', 'Submit').click();
        cy.wait('@createConfiguredTask').its('request.body').should(body => {
            expect(body.subsampleEventCount).to.eq(12345);
            expect(body.configuration.model.architecture).to.eq('mlp');
            expect(body.configuration.model.mlp.dropout).to.eq(0.25);
            expect(body.configuration.model).not.to.have.property('attention');
        });
        cy.contains('a', taskName).click();
        cy.contains('a', 'Reuse in new task').click();
        cy.get('#name').should('have.value', taskName + ' (copy)');
        cy.get('#trainingDatasetId').should('have.value', '1');
        cy.get('#enableSubsampleEventCount').should('be.checked');
        cy.get('#subsampleEventCountInput').should('have.value', '12345');
        cy.contains('summary', /^Model$/).click();
        cy.get(architecture).should('have.value', 'mlp');
        cy.get(dropout).should('be.visible').and('have.value', '0.25');
    });

    it('reveals invalid fields in collapsed nested groups on submit', () => {
        cy.contains('main a', 'Create Training Task').click();

        cy.get('#name').type(generateUniqueName('invalid'));
        cy.get('#trainingDatasetId').select('1');
        cy.contains('summary', /^Model$/).click();
        const dropout = '[name="configuration.model.attention.dropout"]';
        cy.get(dropout).clear().type('2');
        cy.contains('summary', /^Attention$/).click();
        cy.contains('summary', /^Model$/).click();

        cy.contains('button', 'Submit').click();
        cy.get(dropout).should('be.visible').and('be.focused').then($field => {
            expect($field[0].validity.rangeOverflow).to.be.true;
        });
        cy.get(dropout).clear().type('0.4');
        cy.contains('summary', /^Model$/).click();
        cy.intercept('POST', '/training-tasks').as('correctedTask');
        cy.contains('button', 'Submit').click();
        cy.wait('@correctedTask').its('response.statusCode').should('be.within', 200, 299);
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

    it('bounds the list and restores filter, sort and page from the URL', () => {
        cy.intercept('GET', '/training-tasks/list*').as('taskPage');
        cy.get('.task-list-filters [name="name"]').type('PAGING fixture');
        cy.get('.task-list-filters [name="status"]').select('Completed');
        cy.get('.task-list-filters [name="sort"]').select('Oldest first');
        cy.get('[name="from"]').type('20/11/2024');
        cy.get('input[aria-label="Start date calendar"]').should('have.value', '2024-11-20');
        cy.get('input[aria-label="End date calendar"]').then(([calendar]) => {
            cy.stub(calendar, 'showPicker').callsFake(() => {
                calendar.value = '2024-11-21';
                calendar.dispatchEvent(new Event('change', { bubbles: true }));
            }).as('openEndCalendar');
        });
        cy.get('button[aria-label="Open end date calendar"]').click();
        cy.get('@openEndCalendar').should('have.been.calledOnce');
        cy.get('[name="to"]').should('have.value', '21/11/2024');
        cy.get('#task-dataset-value').click();
        cy.contains('#task-dataset-choices button', 'Mixed periods 2023').click();
        cy.contains('button', 'Apply').click();
        cy.wait('@taskPage').its('response.body').should(body => {
            expect(body.length).to.be.lessThan(30000);
        });
        cy.get('.task-list tbody tr').should('have.length', 25);
        cy.get('.task-name').first().should('have.text', 'Paging fixture 0001');
        cy.get('.task-name').last().should('have.text', 'Paging fixture 0025');
        cy.contains('a', 'Next').click();
        cy.location('search').should('include', 'page=2').and('include', 'dataset=1');
        cy.get('.task-list tbody tr').should('have.length', 25);
        cy.get('.task-name').first().should('have.text', 'Paging fixture 0026');

        cy.reload();
        cy.get('.task-name').first().should('have.text', 'Paging fixture 0026');
        cy.get('.task-list-filters [name="name"]').should('have.value', 'PAGING fixture');
        cy.get('.task-list-filters [name="status"]').should('have.value', '4');
        cy.get('.task-list-filters [name="sort"]').should('have.value', 'oldest');
        cy.get('[name="from"]').should('have.value', '20/11/2024');
        cy.get('[name="to"]').should('have.value', '21/11/2024');
        cy.get('.task-list tbody time').first().should('contain.text', '20/11/2024');
        cy.go('back');
        cy.get('.task-name').first().should('have.text', 'Paging fixture 0001');
        cy.go('forward');
        cy.get('.task-name').first().should('have.text', 'Paging fixture 0026');

        cy.get('.task-list-filters [name="sort"]').select('Name (Z–A)');
        cy.contains('button', 'Apply').click();
        cy.location('search').should('not.include', 'page=');
        cy.get('.task-name').first().should('have.text', 'Paging fixture 1000');
    });

    it('uses searchable dropdowns for the dataset and creator filters', () => {
        cy.intercept('GET', '/training-tasks/filter-choices/dataset*').as('datasetChoices');
        cy.get('#task-dataset-search').should('not.be.visible');
        cy.get('#task-dataset-value').click();
        cy.get('#task-dataset-search').should('be.focused').type('LHC23C1');
        cy.wait('@datasetChoices');
        cy.get('#task-dataset-choices button').should('have.length', 2);
        cy.contains('#task-dataset-choices button', 'LHC23c1').click();
        cy.get('#task-dataset-value').should('have.text', 'LHC23c1');
        cy.get('#task-dataset-search').should('not.be.visible');
        cy.get('#task-dataset-value').click();
        cy.get('#task-dataset-search').clear().type('No matching fixture');
        cy.contains('[role="status"]', 'No matches.').should('be.visible');
        cy.get('#task-dataset').should('have.value', '2');
        cy.get('#task-dataset-search').trigger('keydown', { key: 'Escape' });
        cy.get('#task-dataset-value').parent().should('be.focused');

        cy.get('#task-owner-value').click();
        cy.get('#task-owner-search').type('NBOHR');
        cy.get('#task-owner-choices button').should('have.length', 2);
        cy.contains('#task-owner-choices button', 'Niels Bohr (nbohr)').click();
        cy.get('#task-owner-search').should('not.be.visible');
        cy.contains('button', 'Apply').click();
        cy.location('search').should('include', 'dataset=2').and('include', 'owner=2').and('not.include', 'page=');
        cy.get('.task-list tbody tr').should('have.length', 1).and('contain.text', 'LHC24b1b Other Task');
        cy.reload();
        cy.get('#task-dataset-value').should('have.text', 'LHC23c1');
        cy.get('#task-owner-value').should('have.text', 'Niels Bohr (nbohr)');
        cy.go('back');
        cy.get('#task-dataset-value').should('have.text', 'All datasets');
        cy.get('#task-owner-value').should('have.text', 'Anyone');
        cy.go('forward');
        cy.get('#task-dataset-value').should('have.text', 'LHC23c1');
        cy.get('#task-owner-value').should('have.text', 'Niels Bohr (nbohr)');
        cy.get('.task-list tbody tr').should('have.length', 1);
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
        cy.contains('button', 'Apply').click();
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
                expect(doc.documentElement.scrollWidth).to.be.at.most(doc.body.clientWidth);
            });
        });
    }

});
