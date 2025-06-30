import {Injectable} from '@angular/core';
import {HttpClient} from "@angular/common/http";
import {Observable} from "rxjs";
import {Pod} from "./terminal";

@Injectable({
  providedIn: 'root'
})
export class TerminalService {

  constructor(private http: HttpClient) {
  }

  createTerminalSession(namespace: string, podName: string, shell: string): Observable<any> {
    const url = function () {
      let baseUrl = `/terminals?podName=${podName}&&shell=${shell}`
      if (namespace) {
        baseUrl = `${baseUrl}&&namespace=${namespace}`
      }
      return baseUrl
    }()
    return this.http.get<any>(url)
  }
}
