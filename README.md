# crawlfish 

a github `.md` 404 crawler

### project goal 

Take in a list of github repos (list of string urls). For each repo...

Crawl to find all `.md` files. Get all hyperlinks from each .md file. 

Make an HTTP request to it and see if response is 404. Keep track of all `not found` pages. Return them. 

