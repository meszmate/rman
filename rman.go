package rman

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"

	"github.com/klauspost/compress/zstd"
	manifest "github.com/meszmate/manifest"
	"github.com/meszmate/rman/flatbuffer"
)


func LoadURLBytes(url string) []byte{
	return manifest.LoadURLBytes(url)
}
func LoadFileBytes(url string) []byte{
	return manifest.LoadFileBytes(url)
}


type Manifest struct {
    ID          uint64
    Flags       []Flag
    Files       []FileEntry
}

type Chunk struct {
    CompressedSize   uint32
    UncompressedSize uint32
    ChunkID          int64
    BundleOffset     uint32
    BundleID         int64
}

type Flag struct {
    FlagID      uint8
    Name        string
}

type FileEntry struct {
    ID              int64
    FileSize        uint32
    Name            string
    SymLink         string
    Flags           []Flag
    Chunks          []Chunk
}

type FirstDirectory struct {
    DirectoryID int64
    ParentID    int64
    Name        string
}
type Directory struct {
    DirectoryID int64
    Name        string
}

func Decompress(data []byte) []byte{
    decoder, err := zstd.NewReader(nil)
    if err != nil {
        log.Fatal(err)
    }
    defer decoder.Close()

    newData, err := decoder.DecodeAll(data, nil)
    if err != nil {
        log.Fatal(err)
    }
    return newData
}

func ParseBody(manifest *Manifest, data []byte) error {

    root := flatbuffer.GetRootAsManifest(data, 0)

    manifest.Flags = make([]Flag, 0)
    FlagsLength := root.TagsLength()
    for i := 0; i < FlagsLength; i++ {
        flag := new(flatbuffer.Tag)
        root.Tags(flag, i)

        manifest.Flags = append(manifest.Flags, Flag{
            Name: string(flag.Name()),
            FlagID: flag.Id(),
        })
    }

    var FirstDirectories []FirstDirectory = make([]FirstDirectory, 0)
    FirstDirectoriesLength := root.DirectoriesLength()
    for i := 0; i < FirstDirectoriesLength; i++ {
        dir := new(flatbuffer.Directory)
        root.Directories(dir, i)
        
        if dir.Id() == 0 {
            continue
        }
        FirstDirectories = append(FirstDirectories, FirstDirectory{
            DirectoryID: dir.Id(),
            ParentID: dir.ParentId(),
            Name: string(dir.Name()),
        })
    }
    var Directories []Directory = make([]Directory, 0)
    for _, k := range FirstDirectories{
        if k.ParentID == 0{
            Directories = append(Directories, Directory{
                DirectoryID: k.DirectoryID,
                Name: k.Name,
            })
            continue
        }
        fullpath := k.Name
        
        ParentID := k.ParentID
        for ParentID != 0{
            for _, d := range FirstDirectories{
                if d.DirectoryID == ParentID{
                    fullpath = d.Name + "/" + fullpath
                    ParentID = d.ParentID
                    break
                }
            }
        }
        Directories = append(Directories, Directory{
            DirectoryID: k.DirectoryID,
            Name: fullpath,
        })
    }
    
    var Chunks map[int64]Chunk = make(map[int64]Chunk, 0)
    BundleLength := root.BundlesLength()
    var LastChunk Chunk
    for i := 0; i < BundleLength; i++{
        bundle := new(flatbuffer.Bundle)
        root.Bundles(bundle, i)

        ChunksLength := bundle.ChunksLength()
        for j := 0; j < ChunksLength; j++{
            chunk := new(flatbuffer.Chunk)
            bundle.Chunks(chunk, j)

            var BundleOffset uint32
            if j == 0{
                BundleOffset = 0
            } else {
                BundleOffset = LastChunk.BundleOffset + LastChunk.CompressedSize
            }
            newChunk := Chunk{
                BundleID: bundle.Id(),
                CompressedSize: chunk.CompressedSize(),
                UncompressedSize: chunk.UncompressedSize(),
                BundleOffset: BundleOffset,
                ChunkID: chunk.Id(),
            }

            LastChunk = newChunk

            Chunks[newChunk.ChunkID] = newChunk
        }
    }
    manifest.Files = make([]FileEntry, 0)

    files := root.FilesLength()
    for i := 0; i < files; i++ {
        file := new(flatbuffer.File)
        root.Files(file, i)
        
        var fileEntry FileEntry

        flags := file.TagBitmask()
        fileEntry.Flags = make([]Flag, 0)
        for _, f := range fileEntry.Flags{
            if flags&(1<<f.FlagID) != 0{
                fileEntry.Flags = append(fileEntry.Flags, f)
            }
        }

        fileEntry.Name = string(file.Name())
        for _, d := range Directories{
            if d.DirectoryID == file.DirectoryId(){
                fileEntry.Name = d.Name + "/" + fileEntry.Name
                break
            }
        }
        fileEntry.SymLink = string(file.Symlink())
        fileEntry.FileSize = file.Size()
        fileEntry.ID = file.Id()

        fileEntry.Chunks = make([]Chunk, 0)
        chunkslength := file.ChunkIdsLength()
        for j := 0; j < chunkslength; j++ {
            var chunkID int64 = file.ChunkIds(j)
            ch, ok := Chunks[chunkID]
            if !ok{
                fmt.Printf("ChunkID %d not found", chunkID)
            }
            fileEntry.Chunks = append(fileEntry.Chunks, ch)
        }
        manifest.Files = append(manifest.Files, fileEntry)
    }

    return nil
}

func ParseManifestData(data []byte) (*Manifest, error) {
    if len(data) < 28 {
        return nil, fmt.Errorf("data too short to contain necessary headers")
    }

    if string(data[:4]) != "RMAN" {
        return nil, fmt.Errorf("Not a valid RMAN file! Missing magic bytes.")
    }

    version := data[4]
    if version == 2 && data[5] != 0 {
        fmt.Printf("Info: Untested manifest version %d.%d detected. Everything should still work though.\n", version, data[5])
    } else if version != 2 {
        fmt.Printf("Warning: Probably unsupported manifest version %d.%d detected. Will continue, but it might not work.\n", version, data[5])
    }

    contentOffset := binary.LittleEndian.Uint32(data[8:12])
    compressedSize := binary.LittleEndian.Uint32(data[12:16])
    uncompressedSize := binary.LittleEndian.Uint32(data[24:28])

    if int(contentOffset+compressedSize) > len(data) {
        return nil, fmt.Errorf("Compressed data exceeds the size of the manifest")
    }

    compressedData := data[contentOffset : contentOffset+compressedSize]

    reader := bytes.NewReader(compressedData)
    decoder, err := zstd.NewReader(reader)
    if err != nil {
        return nil, fmt.Errorf("Zstandard decompression failed: %w", err)
    }
    defer decoder.Close()

    uncompressedData, err := io.ReadAll(decoder)
    if err != nil {
        return nil, fmt.Errorf("Reading decompressed data failed: %w", err)
    }

    if len(uncompressedData) != int(uncompressedSize) {
        return nil, fmt.Errorf("Decompressed data size mismatch: expected %d, got %d", uncompressedSize, len(uncompressedData))
    }

    manifest := &Manifest{}
    manifest.ID = binary.LittleEndian.Uint64(data[16:24])
    if err := ParseBody(manifest, uncompressedData); err != nil {
        return nil, fmt.Errorf("Error parsing body: %w", err)
    }

    return manifest, nil
}
