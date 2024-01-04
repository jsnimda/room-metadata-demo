import { HttpClient } from '@angular/common/http';
import { Component, inject, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { Room, RoomEvent } from 'livekit-client';
import { BehaviorSubject, map } from 'rxjs';

const apiUrl = 'http://localhost:8101';
const livekitUrl = 'ws://localhost:7880';

interface CreateRoomResponse {
  roomName: string;
  token: string;
}

interface RoomMetadata {
  counter: number;
}

@Component({
  selector: 'app-folder',
  templateUrl: './folder.page.html',
  styleUrls: ['./folder.page.scss'],
})
export class FolderPage implements OnInit {
  public folder!: string;
  private activatedRoute = inject(ActivatedRoute);
  constructor(
    private http: HttpClient,
  ) {}

  messages: string[] = [];
  room: Room = new Room();
  metadata = new BehaviorSubject<RoomMetadata>({ counter: 0 });
  counter = this.metadata.pipe(map((m) => m.counter));

  private logMessage(msg: string) {
    console.log(msg);
    this.messages.push(`log: ${msg}`);
  }

  private errorMessage(msg: string) {
    console.error(msg);
    this.messages.push(`error: ${msg}`);
  }

  ngOnInit() {
    this.folder = this.activatedRoute.snapshot.paramMap.get('id') as string;

    this.room.on(RoomEvent.RoomMetadataChanged, (metadata) => {
      this.logMessage(`RoomEvent.RoomMetadataChanged: ${JSON.stringify(metadata)}`);
      if (metadata) {
        this.metadata.next(JSON.parse(metadata));
      }
    });

    this.http.post<CreateRoomResponse>(`${apiUrl}/create-room`, {}).subscribe({
      next: (res) => {
        this.logMessage(`/create-room: ${res.roomName}`);
        this.joinRoom(res.roomName, res.token);
      },
      error: (err) => {
        console.error(err);
        this.errorMessage('failed to create room');
      },
    });
  }

  private async joinRoom(_: string, token: string) {
    this.room.on(RoomEvent.RoomMetadataChanged, (metadata) => {
      this.logMessage(`[Type0] RoomEvent.RoomMetadataChanged: ${JSON.stringify(metadata)}`);
      if (metadata) {
        this.metadata.next(JSON.parse(metadata));
      }
    });

    await this.room.connect(livekitUrl, token);

    this.room.on(RoomEvent.RoomMetadataChanged, (metadata) => {
      this.logMessage(`[Type1] RoomEvent.RoomMetadataChanged: ${JSON.stringify(metadata)}`);
      if (metadata) {
        this.metadata.next(JSON.parse(metadata));
      }
    });

    this.logMessage(`joined room ${this.room.name}, this.room.metadata: ${this.room.metadata}`);
  }

  public async increment() {
    this.http.post(`${apiUrl}/counter-increment`, {
      roomName: this.room.name,
    }).subscribe();

    setTimeout(() => {
      this.logMessage(`after increment() 1s, this.room.metadata: ${this.room.metadata}`);
    }, 1000);
  }
}
