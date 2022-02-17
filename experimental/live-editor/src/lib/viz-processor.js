export const toMermaid = obj => {
    let lines = [
        "flowchart RL"
    ]
    if (obj.spec.resources) {
        obj.spec.resources.forEach(resource => {
            if (!resource.name) {
                return
            }
            lines.push(`res-${resource.name}["${resource.name}"]`);
            ["sources", "images", "configs"].forEach(resourceType => {
                if (resource[resourceType]) {
                    resource[resourceType].forEach(input => {
                        lines.push(`res-${resource.name} --> res-${input.resource}`)
                    })
                }
            })

        })
    }
    // console.log(lines)
    return lines.join("\n")
}