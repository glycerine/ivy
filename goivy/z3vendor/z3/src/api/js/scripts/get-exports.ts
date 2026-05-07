import { functions } from './parse-api';

const extras = [
  '_malloc',
  '_free',
];

const fns = functions.map(f => '_' + f.name);
console.log(JSON.stringify([...extras, ...fns]));
