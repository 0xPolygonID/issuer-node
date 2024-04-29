import * as dotenv from 'dotenv';
import { test, expect } from './fixtures';
import * as metamask from "@synthetixio/synpress/commands/metamask";
import { formatDate } from '../src/utils/forms'


test.beforeEach(async ({ page }) => {
  // baseUrl is set in playwright.config.ts
  await page.goto("https://schema-builder-dev.polygonid.me");
});

test('Import schema to issuer node', async ({ page }) => {
  var currentDate = formatDate(new Date(), 'date-time');
  await page.goto('https://schema-builder-dev.polygonid.me');
  await page.bringToFront();
  await page.getByRole('menuitem', { name: 'Schema Builder' }).click();
  await page.getByRole('button', { name: 'Accept' }).click();
  await page.getByText("Connect wallet").click();
  await metamask.acceptAccess({signInSignature: true});
  await page.getByPlaceholder('Enter a title').fill(`TestTitle ${currentDate}`);
  await page.getByPlaceholder('Enter type').fill(`TestSchemaType`);
  await page.locator('#version').fill(`Version ${currentDate}`);
  await page.getByLabel('Description').fill('Description');
  await page.getByRole('button', { name: 'Define attributes' }).click();
  await page.getByRole('button', { name: 'Add' }).click();
  await page.getByPlaceholder('Enter name').fill('String');
  await page.getByPlaceholder('Enter title').fill('String');
  await page.getByLabel('Description').fill('Desc');
  await page.getByText('credentialSubject', { exact: true }).click();
  await page.getByRole('button', { name: 'Add' }).click();
  await page.getByPlaceholder('Enter name').fill('Int');
  await page.getByPlaceholder('Enter title').fill('Int');
  await page.locator('.ant-select-selection-item').last().click();
  await page.getByTitle('integer').locator('div').click();
  await page.getByLabel('Description').fill('Desc');
  await page.getByText('credentialSubject', { exact: true }).click();
  await page.getByRole('button', { name: 'Add' }).click();
  await page.getByPlaceholder('Enter name').fill('Numb');
  await page.getByPlaceholder('Enter title').fill('Numb');
  await page.locator('.ant-select-selection-item').last().click();
  await page.getByTitle('number').locator('div').click();
  await page.getByLabel('Description').fill('Desc');
  await page.getByText('credentialSubject', { exact: true }).click();
  await page.getByRole('button', { name: 'Add' }).click();
  await page.getByPlaceholder('Enter name').fill('Bool');
  await page.getByPlaceholder('Enter title').fill('Bool');
  await page.locator('.ant-select-selection-item').last().click();
  await page.getByTitle('boolean').locator('div').click();
  await page.getByLabel('Description').fill('Desc');
  await page.getByRole('button', { name: 'Publish on IPFS' }).click();
  await page.locator('.ant-modal-body .ant-form-item-control-input').click();
  await page.getByText('Age').click();
  await page.getByRole('button', { name: 'Publish on IPFS' }).last().click();
  await expect(page.locator('h2.ant-typography').first()).toHaveText(`TestTitle ${currentDate}`);
  var jsonLd = await page.locator('a').nth(4).getAttribute('href');
  var json = await page.locator('a').nth(3).getAttribute('href');

  dotenv.config({path: "../.env-ui"});
  await page.goto(`https://${process.env.ISSUER_UI_AUTH_USERNAME}:${process.env.ISSUER_UI_AUTH_PASSWORD}@issuer-ui-testing-testnet.polygonid.me/`);

  await page.getByRole('button', { name: 'Import schema' }).click();
  if (json != null) {
    await page.getByPlaceholder('Enter URL').fill(json);
  } else {
    expect(true).toBeFalsy();
  }
  await page.getByRole('button', { name: 'Fetch' }).click();
  await expect(page.locator("anticon-spin")).toHaveCount(0, {timeout: 300000});
  await page.getByRole('button', { name: 'Preview import' }).click();
  await page.getByRole('button', { name: 'Import' }).click();
  await expect(page.getByText(`Version ${currentDate}`, { exact: true })).toBeVisible();
  await page.locator('a').nth(6).click();
  await page.locator('.ant-radio').nth(1).click();
  await page.getByRole('button', { name: 'Next step' }).click();
  await page.getByPlaceholder('Type string').fill('String');
  await page.getByPlaceholder('Type integer').fill('2');
  await page.getByPlaceholder('Type number').fill('4');
  await page.getByRole('button', { name: 'Create credential link' }).click();
  console.log("STOP")
});