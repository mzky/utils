// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sha1

import "github.com/mzky/utils/internal/cpu"

var useAsm = cpu.S390X.HasSHA1
