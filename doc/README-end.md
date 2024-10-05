```

## Sharing presentations

- enter presentation folder
- type `present -c` - present.tar.gz will be created
- send file
  - user can start presentation with `present -f present.tar.gz`
  - user can unpack file, enter folder and execute `present`

## Customizations & security

- present will watch for three files:
  - `present.env` in active directory (this will not override already existing ENV values)
  - `.env` in active directory (this **will** override already existing ENV values)
  - `present.env` in user HOME directory (this **will** override already existing ENV values)
- `.env` file or corresponding variables can be used to customize behavior
  ```txt
  ADMIN_PWD=AdminPassword123
  USER_PWD=user
  ASK_USERNAME=false
  PORT=8080
  NEXT_PAGE=ArrowRight,ArrowDown,PageDown,Space,e
  PREVIOUS_PAGE=ArrowLeft,ArrowUp,PageUp
  TERMINAL_CAST=r,b
  TERMINAL_CLOSE=c
  MENU=m
  ```
- if `ADMIN_PWD` is set, only users authorized with that password can execute the code
  - if `ADMIN_PWD` is not provided, it will be generated and written on console
    - with `ADMIN_PWD_DISABLE=true` you can remove need for admin password
  - admin password can be entered on `/login` url (for example http://localhost:8080/login)
- if `USER_PWD` is set, all 'users' will need to enter password to see the presentation
- rest are pretty self explanatory (also in examples are defaults for all options)

## IDE - syntax highlighter & code snippets

- syntax highlighting is both available for VS Code and VSCodium

  - [Visual Studio Code Marketplace](https://marketplace.visualstudio.com/items?itemName=ZlatkoBratkovic.vscode-oktalz-present)

  - [Open VSX Marketplace](https://open-vsx.org/extension/ZlatkoBratkovic/vscode-oktalz-present)
