/**
 * generate timestamp for each production build
 */
const fs = require('fs');
const data = fs.readFileSync('src/environments/environment.prod.ts', 'utf8').split('\n');

let buildTimestampIndex = 0;
data.forEach((item,index) => {
  if(item.indexOf('buildTimestamp') !== -1) {
    buildTimestampIndex = index;
  }
});
if (buildTimestampIndex > 0) {
  const timestamp = new Date().getTime();
  data[buildTimestampIndex] = `  buildTimestamp: ${timestamp},`;
  fs.writeFileSync('src/environments/environment.prod.ts', data.join('\n'), 'utf8');
}


