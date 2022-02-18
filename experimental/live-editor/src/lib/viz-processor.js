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