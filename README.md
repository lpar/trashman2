# TrashMan

This is a new Go implementation of [TrashMan][1]. It should be portable to
the same places as the C version, as they are all supported by the [xattr][2]
package.

Basically, I operate the folder equivalent of Inbox Zero. If something I
download or create is important, I'll file it away in its proper place.
If I haven't done so within a week or two, it's obviously crap and can be
deleted. So I have a program do so automatically, without even asking me.

Moved to https://codeberg.org/meta/trashman

[1]: https://codeberg.org/meta/trashman
[2]: http://github.com/pkg/xattr
