// web/app/e2e/core-flows.spec.ts — primary-flow suite (RFC-010 K.4; RFC-011 J.4 e2e tier).
// Runs against ephemeral stack (make dev-up + tools/seed fixtures).
import { test, expect } from '@playwright/test';

test.describe('onboarding', () => {
  test('register cluster -> helm command -> live connect flip', async ({ page }) => {
    await page.goto('/admin/clusters');
    await page.getByRole('button', { name: 'Register cluster' }).click();
    await page.getByLabel('Display name').fill('lab-cluster');
    await page.getByRole('button', { name: 'Create' }).click();
    await expect(page.getByTestId('helm-command')).toContainText('helm install nydux-collector');
    await page.request.post('/e2e/simulate/collector-heartbeat', { data: { cluster: 'lab-cluster' } });
    await expect(page.getByTestId('cluster-status-lab-cluster')).toHaveText('connected', { timeout: 30_000 });
  });
});

test.describe('kernel explorer', () => {
  test('filter by KES, open detail, components sum-explain score', async ({ page }) => {
    await page.goto('/kernels');
    await page.getByPlaceholder('filter').fill('kes<50;arch==sm_90');
    await page.keyboard.press('Enter');
    await page.getByRole('row').nth(1).click();
    await expect(page.getByTestId('kes-radar')).toBeVisible();
    await expect(page.getByTestId('kes-components')).toContainText(/roof|occ|stall|mem|tc|mix/);
    await page.getByRole('tab', { name: 'IR' }).click();
    await expect(page.getByTestId('ir-privacy-explainer')).toBeVisible(); // SaaS: features-only
  });
});

test.describe('recommendation lifecycle', () => {
  test('approve requires rationale; token gates apply; rollback visible', async ({ page }) => {
    await page.goto('/recommendations');
    await page.getByRole('row', { name: /unfused_elementwise/ }).click();
    await page.getByRole('button', { name: 'Approve' }).click();
    await page.getByRole('button', { name: 'Confirm' }).click();
    await expect(page.getByText('Rationale is required')).toBeVisible();
    await page.getByLabel('Rationale').fill('verified gain on staging');
    await page.getByRole('button', { name: 'Confirm' }).click();
    await expect(page.getByTestId('rec-state')).toHaveText('approved');
    await page.getByRole('button', { name: 'Apply' }).click();
    await expect(page.getByTestId('rec-state')).toHaveText('applied');
    await expect(page.getByTestId('rollback-button')).toBeEnabled();
  });
  test('separation of duties: author cannot approve own rec', async ({ page }) => {
    await page.goto('/recommendations/00000000-0000-7000-8000-00000000e2e1');
    await expect(page.getByRole('button', { name: 'Approve' })).toBeDisabled();
    await expect(page.getByTestId('sod-explainer')).toContainText('author');
  });
});

test.describe('CI gate + regression matrix', () => {
  test('pre-upgrade risk report renders; snippet copies exact CLI gate', async ({ page, context }) => {
    await context.grantPermissions(['clipboard-read', 'clipboard-write']);
    await page.goto('/compilers/regressions');
    await page.getByLabel('From toolchain').selectOption({ index: 1 });
    await page.getByLabel('To toolchain').selectOption({ index: 2 });
    await page.getByRole('button', { name: 'Run risk report' }).click();
    await expect(page.getByTestId('cri-heatmap')).toBeVisible({ timeout: 60_000 });
    await page.getByRole('button', { name: 'Copy CI snippet' }).click();
    const clip = await page.evaluate(() => navigator.clipboard.readText());
    expect(clip).toContain('nydux regressions --fail-on CRI>0.10');
  });
});

test.describe('savings re-anchor co-sign', () => {
  test('wizard requires both signatures and links audit entry', async ({ page }) => {
    await page.goto('/savings');
    await page.getByRole('button', { name: 'Re-anchor baseline' }).click();
    await page.getByLabel('Reason').fill('model architecture change 2026-07');
    await page.getByRole('button', { name: 'Customer sign' }).click();
    await expect(page.getByRole('button', { name: 'Finish' })).toBeDisabled();
    await page.getByRole('button', { name: 'Request NYDUX co-sign' }).click();
    await page.request.post('/e2e/simulate/nydux-cosign');
    await page.getByRole('button', { name: 'Finish' }).click();
    await expect(page.getByTestId('baseline-version')).not.toHaveText('1');
    await page.getByTestId('audit-link').click();
    await expect(page).toHaveURL(/governance\/audit/);
  });
});

test.describe('accessibility', () => {
  test('dashboard passes axe AA', async ({ page }) => {
    await page.goto('/');
    const { injectAxe, checkA11y } = await import('axe-playwright');
    await injectAxe(page);
    await checkA11y(page);
  });
});
