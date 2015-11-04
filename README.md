Go Bintray
==========

[![Build Status](https://travis-ci.org/enr/go-bintray.png?branch=master)](https://travis-ci.org/enr/go-bintray)
[![Build status](https://ci.appveyor.com/api/projects/status/hqiotpa8gqt25bhy?svg=true)](https://ci.appveyor.com/project/enr/go-bintray)

Go library for accessing [Bintray](https://bintray.com/) API.

Import the library:

```Go
    import (
        "github.com/enr/go-bintray/bintray"
    )
```

**Create client**

API:

```Go
    NewClient(httpClient *http.Client, subject, apikey string) *BintrayClient
```

Example:

```Go
    c := NewClient(nil, "subject", "apikey")
```

**Package exists**

API:

```Go
    PackageExists(subject, repository, pkg string) (bool, error)
```

Example:

```Go
    pe, err := client.PackageExists("subject", "repository", "pkg")
    if err != nil {
        t.Errorf("unexpected error thrown %s", err)
    }
```

**Get versions**

API:

```Go
    GetVersions(subject, repository, pkg string) ([]string, error)
```

Example:

```Go
    versions, err := client.GetVersions("subject", "repository", "pkg")
    if err != nil {
        t.Errorf("unexpected error thrown %s", err)
    }
```

**Create version**

API:

```Go
    CreateVersion(subject, repository, pkg, version string) error
```

Example:

```Go
    err := client.CreateVersion("subject", "repository", "pkg", "0.1.2")
    if err != nil {
        t.Errorf("unexpected error thrown %s", err)
    }
```

**Upload file**

API:

```Go
    UploadFile(subject, repository, pkg, version, projectGroupId, projectName, filePath, extraArgs string, mavenRepo bool) error
```

Example:

```Go
    err := client.UploadFile("subject", "repository", "pkg", "1.2", "", "", "testdata/01.txt", "", false)
    if err != nil {
        t.Errorf("unexpected error thrown %s", err)
    }
```

**Publish file**

API:

```Go
    Publish(subject, repository, pkg, version string) error
```

Example:

```Go
    err := client.Publish("subject", "repository", "pkg", "1.2")
    if err != nil {
        t.Errorf("unexpected error thrown %s", err)
    }
```


License
-------

Apache 2.0 - see LICENSE file.

   Copyright 2014 go-bintray contributors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
