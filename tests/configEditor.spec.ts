import { test, expect } from '@grafana/plugin-e2e';
import { MyDataSourceOptions, MySecureJsonData } from '../src/types';

test('smoke: should render config editor', async ({ createDataSourceConfigPage, readProvisionedDataSource, page }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await createDataSourceConfigPage({ type: ds.type });
  await expect(page.getByRole('textbox', { name: 'Host URL' })).toBeVisible();
});

test('"Save & test" should be successful when configuration is valid', async ({
  createDataSourceConfigPage,
  readProvisionedDataSource,
  page,
}) => {
  const ds = await readProvisionedDataSource<MyDataSourceOptions, MySecureJsonData>({ fileName: 'datasources.yml' });
  const configPage = await createDataSourceConfigPage({ type: ds.type });
  await page.getByTestId('host-input').fill(ds.jsonData.host ?? 'localhost');
  await page.getByTestId('port-input').fill(ds.jsonData.port?.toString() ?? '5432');
  await page.getByTestId('database-input').fill(ds.jsonData.database ?? 'defaultdb');
  await page.getByTestId('username-input').fill(ds.jsonData.username ?? 'root');
  // await page.getByLabel('Host URL').fill(ds.jsonData.host ?? 'localhost');
  // await page.getByPlaceholder('26257').fill(ds.jsonData.port?.toString() ?? '26257');
  // await page.getByPlaceholder('defaultdb').fill(ds.jsonData.database ?? 'defaultdb');
  // await page.getByPlaceholder('root').fill(ds.jsonData.username ?? 'root');
  await page.getByTestId('password-input').fill(ds.secureJsonData?.password ?? '');
  await expect(configPage.saveAndTest()).toBeOK();
});

test('"Save & test" should fail when configuration is invalid', async ({
  createDataSourceConfigPage,
  readProvisionedDataSource,
  page,
}) => {
  const ds = await readProvisionedDataSource<MyDataSourceOptions, MySecureJsonData>({ fileName: 'datasources.yml' });
  const configPage = await createDataSourceConfigPage({ type: ds.type });
  await page.getByLabel('Host URL').fill('');
  await expect(page.getByText('The host address cannot be empty')).toBeVisible();
});
