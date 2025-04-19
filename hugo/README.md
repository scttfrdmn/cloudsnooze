# CloudSnooze GitHub Pages

This directory contains the Hugo-based GitHub Pages site for CloudSnooze.

## Structure

- `config.toml` - Hugo configuration file
- `content/` - Website content (Markdown files)
  - `_index.md` - Homepage content
  - `docs/` - Documentation copied from the main docs folder
- `static/` - Static assets like images and CSS
- `layouts/` - Custom layout templates (if needed)

## Development

To work on the site locally:

1. Install Hugo (extended version): https://gohugo.io/installation/
2. Run: `cd hugo && hugo mod get github.com/google/docsy`
3. Run the site locally: `hugo server -D`
4. Open your browser at http://localhost:1313/

## Deployment

The site is automatically deployed to GitHub Pages when changes are pushed to the main branch, using the GitHub Actions workflow defined in `.github/workflows/github-pages.yml`.