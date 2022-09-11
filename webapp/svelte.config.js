import sveltePreprocess from 'svelte-preprocess'
import makeAttractionsImporter from 'attractions/importer.js';
import path from 'node:path';

export default {
  // Consult https://github.com/sveltejs/svelte-preprocess
  // for more information about preprocessors
  preprocess: sveltePreprocess({
    scss: {
      importer: makeAttractionsImporter({
        // specify the path to your theme file, relative to this file
        themeFile: 'src/theme.scss',
      }),
    },
  })
}
