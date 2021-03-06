package metadata;

// Read and write users from streams.

import (
    "bufio"
    "encoding/json"
    "fmt"
    "io"

    "github.com/pkg/errors"

    "github.com/eriq-augustine/elfs/cipherio"
    "github.com/eriq-augustine/elfs/identity"
    "github.com/eriq-augustine/elfs/util"
)

// Read all users into memory.
// This function will not clear the given users.
// However, the reader WILL be closed.
func ReadUsers(users map[identity.UserId]*identity.User, reader util.ReadSeekCloser) (int, error) {
    version, err := ReadUsersWithScanner(users, bufio.NewScanner(reader));
    if (err != nil) {
        return 0, errors.WithStack(err);
    }

    return version, errors.WithStack(reader.Close());
}
func ReadUsersWithScanner(users map[identity.UserId]*identity.User, scanner *bufio.Scanner) (int, error) {
    size, version, err := scanMetadata(scanner);
    if (err != nil) {
        return 0, errors.WithStack(err);
    }

    // Read all the users.
    for i := 0; i < size; i++ {
        var entry identity.User;

        if (!scanner.Scan()) {
            err = scanner.Err();

            if (err == nil) {
                return 0, errors.Wrapf(io.EOF, "Early end of Users. Only read %d of %d entries.", i , size);
            } else {
                return 0, errors.Wrapf(err, "Bad scan on Users entry %d.", i);
            }
        }

        err = json.Unmarshal(scanner.Bytes(), &entry);
        if (err != nil) {
            return 0, errors.Wrapf(err, "Error unmarshaling the user at index %d (%s).", i, string(scanner.Bytes()));
        }

        users[entry.Id] = &entry;
    }

    return version, nil;
}

// Write all users.
// This function will not close the given writer.
func WriteUsers(users map[identity.UserId]*identity.User, version int, writer *cipherio.CipherWriter) error {
    err := writeMetadata(writer, len(users), version);
    if (err != nil) {
        return errors.WithStack(err);
    }

    // Write all the users.
    for i, entry := range(users) {
        line, err := json.Marshal(entry);
        if (err != nil) {
            return errors.Wrapf(err, "Failed to marshal User entry %d.", i);
        }

        _, err = writer.Write([]byte(fmt.Sprintf("%s\n", string(line))));
        if (err != nil) {
            return errors.Wrapf(err, "Failed to write User entry %d.", i);
        }
    }

    return nil;
}
