# (WIP) Symmetric encryption and decryption layer for files

This Mattermost plugin serves as a symmetric encryption and decryption layer when uploading and downloading files, respectively.

This is a work in progress and contributions are welcome!

This plugin needs a hook that is invoked when a file will be read. There is no such hook for server-side plugins in Mattermost server (as of 29.11.2019). So, I am currently working on adding a new hook named **FileWillBeRead**. It is a work-in-progress and can be seen [here](https://github.com/HilalNazli/mattermost-server/tree/add-FileWillBeRead-hook-for-server-side-plugins). Comments and contributions to there are also welcome!