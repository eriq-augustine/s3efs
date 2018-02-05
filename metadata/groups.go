package metadata;

// Read and write groups from streams.

import (
   "bufio"
   "encoding/json"
   "fmt"
   "io"

   "github.com/pkg/errors"

   "github.com/eriq-augustine/elfs/cipherio"
   "github.com/eriq-augustine/elfs/group"
)

// Read all groups into memory.
// This function will not clear the given groups.
// However, the reader WILL be closed.
func ReadGroups(groups map[group.Id]*group.Group, reader cipherio.ReadSeekCloser) (int, error) {
   version, err := ReadGroupsWithScanner(groups, bufio.NewScanner(reader));
   if (err != nil) {
      return 0, errors.WithStack(err);
   }

   return version, errors.WithStack(reader.Close());
}

// Same as the other read, but we will read directly from a deocder
// owned by someone else.
// This is expecially useful if there are multiple
// sections of metadata written to the same file.
func ReadGroupsWithScanner(groups map[group.Id]*group.Group, scanner *bufio.Scanner) (int, error) {
   size, version, err := scanMetadata(scanner);
   if (err != nil) {
      return 0, errors.WithStack(err);
   }

   // Read all the groups.
   for i := 0; i < size; i++ {
      var entry group.Group;

      if (!scanner.Scan()) {
         err = scanner.Err();

         if (err == nil) {
            return 0, errors.Wrapf(io.EOF, "Early end of Groups. Only read %d of %d entries.", i , size);
         } else {
            return 0, errors.Wrapf(err, "Bad scan on Groups entry %d.", i);
         }
      }

      err = json.Unmarshal(scanner.Bytes(), &entry);
      if (err != nil) {
         return 0, errors.Wrapf(err, "Error unmarshaling the group at index %d (%s).", i, string(scanner.Bytes()));
      }

      groups[entry.Id] = &entry;
   }

   return version, nil;
}

// Write all groups.
// This function will not close the given writer.
func WriteGroups(groups map[group.Id]*group.Group, version int, writer *cipherio.CipherWriter) error {
   var bufWriter *bufio.Writer = bufio.NewWriter(writer);

   err := writeMetadata(bufWriter, len(groups), version);
   if (err != nil) {
      return errors.WithStack(err);
   }

   // Write all the groups.
   for i, entry := range(groups) {
      line, err := json.Marshal(entry);
      if (err != nil) {
         return errors.Wrapf(err, "Failed to marshal Group entry %d.", i);
      }

      _, err = bufWriter.WriteString(fmt.Sprintf("%s\n", string(line)));
      if (err != nil) {
         return errors.Wrapf(err, "Failed to write Group entry %d.", i);
      }
   }

   return errors.WithStack(bufWriter.Flush());
}
