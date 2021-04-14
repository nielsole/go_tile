The test file is coming from /8/0/0/0/133/128.meta

https://github.com/openstreetmap/mod_tile/blob/master/src/store_file.c#L76

```
	meta_offset = xyzo_to_meta(path, sizeof(path), store->storage_ctx, xmlconfig, options, x, y, z);
```

https://github.com/openstreetmap/mod_tile/blob/9b6124790ae732f25f0dc71cba81fcb565716a05/includes/metatile.h#L38
```
struct entry {
	int offset;
	int size;
};

struct meta_layout {
	char magic[4];
	int count; // METATILE ^ 2
	int x, y, z; // lowest x,y of this metatile, plus z
	struct entry index[]; // count entries
	// Followed by the tile data
	// The index offsets are measured from the start of the file
};
```

https://github.com/openstreetmap/mod_tile/blob/master/src/store_file_utils.c
```
// Returns the path to the meta-tile and the offset within the meta-tile
int xyzo_to_meta(char *path, size_t len, const char *tile_dir, const char *xmlconfig, const char *options, int x, int y, int z)
{
	unsigned char i, hash[5], offset, mask;

	// Each meta tile winds up in its own file, with several in each leaf directory
	// the .meta tile name is beasd on the sub-tile at (0,0)
	mask = METATILE - 1;
	offset = (x & mask) * METATILE + (y & mask);
	x &= ~mask;
	y &= ~mask;

	for (i = 0; i < 5; i++) {
		hash[i] = ((x & 0x0f) << 4) | (y & 0x0f);
		x >>= 4;
		y >>= 4;
	}

#ifdef DIRECTORY_HASH

	if (strlen(options)) {
		snprintf(path, len, "%s/%s/%d/%u/%u/%u/%u/%u.%s.meta", tile_dir, xmlconfig, z, hash[4], hash[3], hash[2], hash[1], hash[0], options);
	} else {
		snprintf(path, len, "%s/%s/%d/%u/%u/%u/%u/%u.meta", tile_dir, xmlconfig, z, hash[4], hash[3], hash[2], hash[1], hash[0]);
	}

#else

	if (strlen(options)) {
		snprintf(path, len, "%s/%s/%d/%u/%u.%s.meta", tile_dir, xmlconfig, z, x, y, options);
	} else {
		snprintf(path, len, "%s/%s/%d/%u/%u.meta", tile_dir, xmlconfig, z, x, y);
	}

#endif
	return offset;
}
#endif
```

https://github.com/openstreetmap/mod_tile/blob/master/src/metatile.cpp
```
// Returns the offset within the meta-tile index table
int metaTile::xyz_to_meta_offset(int x, int y, int z)
{
	unsigned char mask = METATILE - 1;
	return (x & mask) * METATILE + (y & mask);
}
```

https://github.com/openstreetmap/mod_tile/blob/master/src/convert_meta.c#L75
```
static void descend(const char *search)
{
	DIR *tiles = opendir(search);
	struct dirent *entry;
	char path[PATH_MAX];

	if (!tiles) {
		//fprintf(stderr, "Unable to open directory: %s\n", search);
		return;
	}

	while ((entry = readdir(tiles))) {
		struct stat b;
		char *p;

		//check_load();

		if (!strcmp(entry->d_name, ".") || !strcmp(entry->d_name, "..")) {
			continue;
		}

		snprintf(path, sizeof(path), "%s/%s", search, entry->d_name);

		if (stat(path, &b)) {
			continue;
		}

		if (S_ISDIR(b.st_mode)) {
			descend(path);
			continue;
		}

		p = strrchr(path, '.');

		if (p) {
			if (unpack) {
				if (!strcmp(p, ".meta")) {
					process_unpack(tile_dir, path);
				}
			} else {
				if (!strcmp(p, ".png")) {
					process_pack(tile_dir, path);
				}
			}
		}
	}

	closedir(tiles);
}
```