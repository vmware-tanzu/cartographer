/**
 * Copyright 2021 VMware
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

export const toMermaid = obj => {
    let lines = [
        "flowchart RL",
        "classDef not-found fill:#f66;"
    ]

    let resourceNodesByName = {}

    if (obj.spec.resources) {
        obj.spec.resources.forEach(resource => {
            if (!resource.name) {
                return
            }
            let nodeName = `res-${resource.name}`
            let nodeLabel = `${resource.name}`

            resourceNodesByName[nodeName] = nodeLabel

            lines.push(`${nodeName}["${nodeLabel}"]`);

            ["sources", "images", "configs"].forEach((resourceType, rowIndex) => {
                if (resource[resourceType]) {
                    resource[resourceType].forEach(input => {
                        if (resourceNodesByName[`res-${input.resource}`]) {
                            lines.push(`${nodeName} --> res-${input.resource}`)
                        } else {
                            let naTarget = `not-found-${input.resource}-${rowIndex}`
                            lines.push(`${nodeName} --> ${naTarget}`)
                            lines.push(`${naTarget}["not-found"]`)
                            lines.push(`class ${naTarget} not-found`)
                        }
                    })
                }
            })

        })
    }
    // console.log(lines)
    return lines.join("\n")
}