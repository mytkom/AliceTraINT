# Updating Docs

This page describes how to work with the documentation that powers the **Docs** section
of AliceTraINT.

The goal is to keep it simple: edit a few markdown files and redeploy the app.

## Where docs live

- All documentation files are stored in the `docs/` directory of the repository.
- Each file is a plain markdown file with the `.md` extension.
- On startup, the application:
  - scans the `docs/` directory,
  - renders each file to HTML,
  - extracts section headings for the sidebar,
  - keeps everything in memory for fast access.

If you change the files, you must restart/redeploy the application for the changes to appear.

## Naming docs and order

- The navigation uses titles taken from the first `#` heading in each file.
- The **URLs** and the **“Other docs”** list on the left are based on the file name:
  - a file `0-getting-started.md` becomes `/docs/getting-started`,
  - a file `2-production-deployment.md` becomes `/docs/production-deployment`.
- A numeric prefix followed by a dash (for example `0-`, `1-`, `2-`) is removed from the URL
  but it is still useful:
  - it lets you control the order of files on disk,
  - it makes it easy to see the intended reading order at a glance.
- The **default page** for `/docs` is the first doc in alphabetical order by title,
  after loading all files.

## Sections shown in the sidebar

The **“On this page”** part of the sidebar is built from headings in the current markdown file:

- the first `#` heading is used as the page title,
- lower‑level headings (`##`, `###`, …) are treated as sections,
- each section gets an automatic anchor, for example:
  - `## Creating a dataset` → a link in the sidebar that scrolls to that section.

To make the sidebar useful:

- write short, clear section titles,
- keep the heading levels consistent (e.g. `##` for main sections, `###` for sub‑sections).

## Linking between docs

You can link from one doc to another using standard markdown links:

- Example:  
  `See [Getting Started](/docs/getting-started) for the basics.`

The path part must match the URL derived from the file name after removing any numeric prefix.

## Images

Images are static files served from the same directory tree as the markdown docs:

- Place images inside the `docs/` directory, ideally in a subdirectory that matches the doc,
  for example:
  - `docs/0-getting-started/training-task-new.png`
- In markdown, reference them through the `/docs/static/` path:

  ```markdown
  ![Create dataset view](/docs/static/0-getting-started/training-task-new.png)
  ```

At runtime this is served directly from the `docs/` directory configured in the app.

## Using HTML inside markdown

Markdown files can also contain explicit HTML. This is useful when you want more control
over formatting than plain markdown gives you.

For example, ordered lists with screenshots sometimes render inconsistently in markdown.
In such cases you can use an HTML list and `<img>` tags to keep numbering stable:

```html
<ol>
  <li>
    Step with an image:
    <p><img src="/docs/static/0-getting-started/example.png" alt="Example"></p>
  </li>
  <li>Next step...</li>
</ol>
```

In general:

- you can mix markdown and HTML in the same file,
- HTML is rendered “as is”, so use it only for trusted documentation content,
- prefer markdown for normal text, and use HTML only when you really need it.

## When changes are visible

Because the application processes docs once at startup, updates are **not** picked up
automatically while the server is running.

To apply changes:

1. Edit or add markdown files in the `docs/` directory.
2. Commit and push your changes (if you use Git for deployment).
3. Restart or redeploy the AliceTraINT application.

After the restart:

- new docs will appear in the sidebar and “Other docs” list,
- updated content and images will be visible,
- changed headings will be reflected in the “On this page” section.
