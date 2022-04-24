
def read_int(metafile):
    #print("Positon: {}".format(metafile.tell()))
    return int.from_bytes(metafile.read(4), "little")

def main():
    with open("../ajt/15/0/66/73/207/8.meta", 'rb') as metafile:
        magic = metafile.read(4)
        print(magic)
        count = read_int(metafile)
        print("Header length: {}".format(count))
        x = read_int(metafile)
        y = read_int(metafile)
        z = read_int(metafile)

        print("X: {}, Y: {}, at Z: {}".format(x,y,z))
        png_files = list()
        for i in range(count):
            offset = read_int(metafile)
            size = read_int(metafile)
            png_files.append((offset, size))
            print("Offset: {}\tSize:{}".format(offset, size))
        for offset, size in png_files:
            with open("output/{}.png".format(offset), "wb") as writefile:
                metafile.seek(offset)
                writefile.write(metafile.read(size))



if __name__ == '__main__':
    main()