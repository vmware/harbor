// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import (
	"fmt"
)

var (
	resourceValueFormats = map[ResourceName]func(int64) string{
		ResourceStorage: byteCountToDisplaySize,
	}
)

func byteCountToDisplaySize(value int64) string {
	const unit = 1024
	if value < unit {
		return fmt.Sprintf("%d B", value)
	}

	div, exp := int64(unit), 0
	for n := value / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %ciB", float64(value)/float64(div), "KMGTPE"[exp])
}
