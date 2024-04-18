# Techdocs: Rewrite relative links

This action's job is to scan through all the Markdown files inside the docs
folder (based on the presence of an `mkdocs.yml` file) and rewrite relative
links that point to files *outside* that docs folder folder to absolute ones.
