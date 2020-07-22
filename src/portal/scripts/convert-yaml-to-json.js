/**
 before execute:
 1.npm install js-yaml --save-dev
 2.npm install ng-swagger-gen --save-dev
 */
//configuration
let inputFile = '../../api/v2.0/swagger.yaml';
const outputDir = 'ng-swagger-gen';

//convert swagger.yaml to swagger.json
const yaml = require('js-yaml');
const fs = require('fs');
if (fs.existsSync('swagger2.yaml')) {
   inputFile = 'swagger2.yaml';
}
if (!fs.existsSync(outputDir)) {
    fs.mkdirSync(outputDir);
}
const swaggerObj = yaml.load(fs.readFileSync(inputFile, {encoding: 'utf-8'}));
// host is not needed as UI uses relative path for back-end APIs
if (swaggerObj.host) {
    delete swaggerObj.host;
}
// enhancement for property 'additionalProperties'
traverseObject(swaggerObj);

fs.writeFileSync(outputDir + '/swagger.json', JSON.stringify(swaggerObj, null, 2));


function traverseObject(obj) {
  if (obj) {
    if (Array.isArray(obj)) {
      for (let i = 0; i < obj.length; i++) {
        traverseObject(obj[i])
      }
    }
    if (typeof obj === 'object') {
      for (let name in obj) {
        if (obj.hasOwnProperty(name)) {
          if (name === 'additionalProperties'
            && obj[name].type === 'object'
            && obj[name].additionalProperties === true) {
            obj[name] = true;
          } else {
            traverseObject(obj[name])
          }
        }
      }
    }
  }
}

