import { Pipe, PipeTransform } from '@angular/core';

@Pipe({
  name: 'replace'
})
export class ReplacePipe implements PipeTransform {

  transform(value: string): string {
    if (value == null) {
      return ""
    }
    return value.replace("_", "");
  }


}