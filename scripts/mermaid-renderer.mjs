import fs from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

import { build } from 'esbuild';
import { chromium } from 'playwright';

const root = path.dirname(fileURLToPath(import.meta.url));
const [input, output = 'diagram.svg'] = process.argv.slice(2);

if (!input) {
    console.error(
        'Usage: node render-mermaid.mjs <input.mmd> [output.svg]'
    );
    process.exit(1);
}

const definition = await fs.readFile(
    path.resolve(input),
    'utf8'
);

// Bundle Mermaid and Devicon for Chromium.
const result = await build({
    stdin: {
        contents: `
      import mermaid from 'mermaid';
      import { icons } from '@iconify-json/devicon';

      mermaid.registerIconPacks([
        {
          name: 'devicon',
          icons,
        },
      ]);

      mermaid.initialize({
        startOnLoad: false,
        securityLevel: 'strict',
        theme: 'default',
      });

      window.renderMermaid = async (definition) => {
        const { svg } = await mermaid.render(
          'mermaid-diagram',
          definition
        );

        return svg;
      };
    `,
        resolveDir: root,
        sourcefile: 'mermaid-renderer.mjs',
        loader: 'js',
    },
    bundle: true,
    platform: 'browser',
    format: 'iife',
    target: 'es2022',
    write: false,
    logLevel: 'silent',
});

const bundledCode = result.outputFiles[0].text;

let browser;

try {
    browser = await chromium.launch({
        headless: true,
    });

    const page = await browser.newPage();

    await page.setContent(`
    <!DOCTYPE html>
    <html>
      <head>
        <meta charset="utf-8">
      </head>
      <body></body>
    </html>
  `);

    await page.addScriptTag({
        content: bundledCode,
    });

    const svg = await page.evaluate(async (source) => {
        if (typeof window.renderMermaid !== 'function') {
            throw new Error('Mermaid renderer not initialized');
        }

        return window.renderMermaid(source);
    }, definition);

    if (!svg.includes('<svg') || !svg.includes('</svg>')) {
        throw new Error('Invalid SVG output');
    }

    await fs.writeFile(
        path.resolve(output),
        svg,
        'utf8'
    );

    console.log(`SVG generated: ${output}`);
} finally {
    await browser?.close();
}