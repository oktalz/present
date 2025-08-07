imagePull = () => {
    fetch('api/cmd/image-pull')
        .then(response => response.text())
        .then(data => {
            console.log(`${data}`)
        })
}
imagePullPython = () => {
    fetch('api/cmd/image-pull-python')
}
