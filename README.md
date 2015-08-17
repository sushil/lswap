## Sample Usage

```bash
sushil@everest ~/temp/testmv
$ tree
.
├── one
│   ├── A -> /Users/sushil/temp/testmv/two/A
│   ├── B -> /Users/sushil/temp/testmv/two/B
│   └── C
└── two
    ├── A
    ├── B
    └── C

2 directories, 6 files

sushil@everest ~/temp/testmv
$ lswap -h
Usage of lswap:
  -contents=[]: comma separated list of contents under source folder to move to target
  -from="": source folder where code currently is
  -to="": target folder where code will be moved

sushil@everest ~/temp/testmv
$ lswap -from one -to two -contents A,B
2015/08/17 08:46:10 A is not a symlink under /Users/sushil/temp/testmv/one

sushil@everest ~/temp/testmv
$ lswap -from two -to one -contents A,B
2015/08/17 08:46:30 source and destination looks good, starting work ..
2015/08/17 08:46:30 done

sushil@everest ~/temp/testmv
$ tree
.
├── one
│   ├── A
│   ├── B
│   └── C
└── two
    ├── A -> /Users/sushil/temp/testmv/one/A
    ├── B -> /Users/sushil/temp/testmv/one/B
    └── C

2 directories, 6 files

sushil@everest ~/temp/testmv
$ tree
.
├── one
│   ├── A
│   ├── B
│   └── C
└── two
    ├── A -> /Users/sushil/temp/testmv/one/A
    ├── B -> /Users/sushil/temp/testmv/one/B
    └── C

2 directories, 6 files

sushil@everest ~/temp/testmv
$ lswap -from two -to one -contents C
2015/08/17 08:47:08 /Users/sushil/temp/testmv/one/C is not a symlink

sushil@everest ~/temp/testmv
$ lswap -from one -to two -contents B
2015/08/17 08:47:29 source and destination looks good, starting work ..
2015/08/17 08:47:29 done

sushil@everest ~/temp/testmv
$ tree
.
├── one
│   ├── A
│   ├── B -> /Users/sushil/temp/testmv/two/B
│   └── C
└── two
    ├── A -> /Users/sushil/temp/testmv/one/A
    ├── B
    └── C
```
