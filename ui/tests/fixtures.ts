import { test as base, chromium, type BrowserContext } from "@playwright/test";
import { initialSetup } from "@synthetixio/synpress/commands/metamask";
import { setExpectInstance } from "@synthetixio/synpress/commands/playwright";
import { resetState } from "@synthetixio/synpress/commands/synpress";
import { prepareMetamask } from "@synthetixio/synpress/helpers";
import dotenv from 'dotenv'; 
import path from 'path'; 

export const test = base.extend<{
  context: BrowserContext;
}>({
  context: async ({}, use) => {
    dotenv.config({path: path.resolve("../.env-ui")});
    // required for synpress as it shares same expect instance as playwright
    await setExpectInstance(expect);

    // download metamask 
    const metamaskPath = await prepareMetamask(
      process.env.METAMASK_VERSION || "10.25.0"
    );

    // prepare browser args
    const browserArgs = [
      `--disable-extensions-except=${metamaskPath}`,
      `--load-extension=${metamaskPath}`,
      "--remote-debugging-port=9222",
    ];

    if (process.env.CI) {
      browserArgs.push("--disable-gpu");
    }

    if (process.env.HEADLESS_MODE) {
      browserArgs.push("--headless=new");
    }

    // launch browser
    const context = await chromium.launchPersistentContext("", {
      headless: false,
      args: browserArgs,
    });

    // wait for metamask
    await context.pages()[0].waitForTimeout(3000);

    // setup metamask
    await initialSetup(chromium, {
      enableExperimentalSettings: true,
      secretWordsOrPrivateKey:
        "fill",
      network: {
        id: 80002,
        network: "Polygon",
        name: "Amoy Testnet",
        nativeCurrency: { name: 'Matic', symbol: 'MATIC', decimals: 18 },
        rpcUrls: {
          default: {
            http: ['https://rpc-amoy.polygon.technology/'],
          }
        },
        blockExplorers: {
          etherscan: {
            name: 'Amoy',
            url: 'https://www.oklink.com/amoy',
          },
        }
      },
      password: "Tester@1234",
      enableAdvancedSettings: true,
    });

    await use(context);

    await context.close();

    await resetState();
  },
});

export const expect = test.expect;