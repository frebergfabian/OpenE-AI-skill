// Ofarligt läs- och navigationstest: ändrar ingenting i sandboxen.
const { test, expect } = require('@playwright/test');

const user = process.env.OPENE_ADMIN_USER || 'sandbox-admin';
const pass = process.env.OPENE_ADMIN_PASSWORD;

// OpenE levererar ISO-8859-1. Läs svar som windows-1252, aldrig med response.text().
async function latin1(response) {
  return new TextDecoder('windows-1252').decode(await response.body());
}

test('katalogen visas utan inloggning', async ({ page }, info) => {
  const res = await page.goto('/');
  expect(res.status()).toBe(200);
  expect(res.headers()['content-type'].toLowerCase()).toContain('iso-8859-1');
  await expect(page.locator('body')).toContainText(/e-tjänst/i);
  await page.screenshot({ path: info.outputPath('katalog.png'), fullPage: true });
});

test('admin loggar in och navigerar i adminsidorna', async ({ page }, info) => {
  test.skip(!pass, 'OPENE_ADMIN_PASSWORD saknas');
  await page.goto('/login');
  await page.locator('input[name="username"]').fill(user);
  await page.locator('input[name="password"]').fill(pass);
  await Promise.all([page.waitForLoadState('load'), page.locator('form [type="submit"]').first().click()]);

  for (const path of ['/flowadmin', '/flowadmin/standardstatuses', '/administration/useradmin', '/flowinstanceadmin']) {
    const res = await page.goto(path);
    expect(res.status(), path).toBe(200);
    expect(await latin1(res), path).not.toMatch(/Ett oväntat fel har inträffat/);
  }
  await page.goto('/flowadmin');
  await expect(page.locator('#lab-main, main, body')).toContainText(/e-tjänst/i);
  await page.screenshot({ path: info.outputPath('flowadmin.png'), fullPage: true });
});

test('oinloggad användare når inte adminsidor', async ({ page }) => {
  await page.goto('/flowadmin');
  await expect(page.locator('input[name="password"]')).toBeVisible();
});
