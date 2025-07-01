import { test, expect } from '@grafana/plugin-e2e';

test('smoke: should render query editor with all elements', async ({ panelEditPage, readProvisionedDataSource, page }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);
  await page.waitForSelector('[aria-label="Query editor row"]', { state: 'visible', timeout: 15000 });

  // Verify query input and buttons are visible
  const queryEditor = panelEditPage.getQueryEditorRow('A');
  const editorContainer = queryEditor.getByTestId('editor-container');
  await expect(editorContainer).toBeVisible({ timeout: 10000 });
  
  // Check for buttons using data-testid
  await expect(queryEditor.getByTestId('run-query-button')).toBeVisible();
  await expect(queryEditor.getByTestId('format-sql-button')).toBeVisible();
});

test('data query should return values', async ({ panelEditPage, readProvisionedDataSource, page }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);
  await page.waitForSelector('[aria-label="Query editor row"]', { state: 'visible', timeout: 15000 });

  // Enter query and run
  const queryEditor = panelEditPage.getQueryEditorRow('A');
  const queryInput = queryEditor.getByTestId('editor-container').getByRole('textbox');
  await queryInput.waitFor({ state: 'visible', timeout: 10000 });
  await queryInput.fill('SELECT value FROM test_data');
  
  // Wait a moment before clicking to ensure the query is properly set
  await page.waitForTimeout(500);
  await queryEditor.getByTestId('run-query-button').click();

  // Wait for query execution and panel update
  await page.waitForTimeout(5000);

  // Verify results in table
  await panelEditPage.setVisualization('Table');
  await page.waitForTimeout(2000);
  
  // Check if data appears in the panel
  await expect(panelEditPage.panel.data).toContainText(['10', '20'], { timeout: 20000 });
});

test('should format SQL query when Format SQL button is clicked', async ({ panelEditPage, readProvisionedDataSource, page }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);
  await page.waitForSelector('[aria-label="Query editor row"]', { state: 'visible', timeout: 15000 });

  const queryEditor = panelEditPage.getQueryEditorRow('A');
  const queryInput = queryEditor.getByTestId('editor-container').getByRole('textbox');

  // Enter unformatted SQL
  await queryInput.fill('select * from table where id=1');

  // Click Format SQL button
  await queryEditor.getByTestId('format-sql-button').click();

  // Verify SQL is formatted
  await expect(queryInput).toHaveValue('select\n  *\nfrom\n  table\nwhere\n  id = 1\n');

});

test('should disable buttons when query text is empty', async ({ panelEditPage, readProvisionedDataSource, page }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);
  await page.waitForSelector('[aria-label="Query editor row"]', { state: 'visible', timeout: 15000 });

  // Check buttons are disabled
  const queryEditor = panelEditPage.getQueryEditorRow('A');
  await expect(queryEditor.getByTestId('run-query-button')).toBeDisabled();
  await expect(queryEditor.getByTestId('format-sql-button')).toBeDisabled();
});
