// set SHA=$(git rev-parse --short HEAD); yarn build-386 && mv betaflight-pid-app.exe dist/gui2-$SHA.exe

var fs = require('fs');

var exec = require('child_process').exec;

var result = function(command, cb){
  var child = exec(command, function(err, stdout, stderr){
    if(err != null){
        return cb(new Error(err), null);
    }else if(typeof(stderr) != "string"){
        return cb(new Error(stderr), null);
    }else{
        return cb(null, stdout);
    }
  });
}

result('git describe --tags --long', function(err, sha) {
  sha = sha.replace(/(.*)-(.*)/, '$1')
  sha = sha.replace(/^\s+|\s+$/g, '')
  console.log(`Building: ${sha}`)

  process.env['VERSION'] = sha

  result('yarn build-386', function(err, response) {
    if (err) {
      console.error(err)
      return
    }

    console.log(response)

    const destExePath = 'dist/gui-' + sha + '.exe'

    fs.rename('betaflight-pid-app.exe', destExePath, function() {
      console.log('Renamed app from betaflight-pid-app.exe to ' + destExePath)

      const versions = [
        {
          version: sha,
          file: destExePath
        }
      ]

      fs.writeFile('dist/versions.json', JSON.stringify(versions), function(err) {
        console.log('Version: dist/versions.json created')

        result(`scp -r dist/. foo.us.to:www/foo.us.to/html/gui2/`, function(err) {
          if (err) {
            console.error(err)
            return
          }

          console.log(`Deployed ${sha} to foo.us.to`)
        })
      })
    })
  })
})