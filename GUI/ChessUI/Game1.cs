using Microsoft.Xna.Framework;
using Microsoft.Xna.Framework.Graphics;
using Microsoft.Xna.Framework.Input;
using System.Net.Sockets;
using System.Text.Json;
using System.Collections.Generic;
using System.Text;
using System;
using System.Diagnostics;
using System.Runtime.InteropServices;
namespace ChessUI;

public class Game1 : Game
{
    private GraphicsDeviceManager _graphics;
    private SpriteBatch _spriteBatch;
    public static GameWindow gw;
  
    StringBuilder myTextBoxDisplayCharacters = new StringBuilder();
    Texture2D backgroundTexture;
    Texture2D ballTexture;
    Texture2D pawnTextureB;
    Texture2D pawnTextureW;
    Texture2D rookTextureB;
    Texture2D rookTextureW;
    Texture2D knightTextureB;
    Texture2D knightTextureW;
    Texture2D bishopTextureB;
    Texture2D bishopTextureW;
    Texture2D queenTextureB;
    Texture2D queenTextureW;
    Texture2D kingTextureB;
    Texture2D kingTextureW;
    Texture2D dotTexture;
    SpriteFont currentFont;
    Piece[][] boardArray;
    
    int[,] legalMoves;

    bool pToMove;
    string wCastleK;
    string wCastleQ;
    string bCastleK;
    string bCastleQ;
    string enPass;
    char[] squareLetter;

    Texture2D[] wTyToText;
    Texture2D[] bTyToText;

    int fullMoves;
    int halfMoves;
    FloatingPiece currentlyHeldPiece;

   // bool canStart;

    public static Process bockFish;

    public static int xOffset = 25;
    public static int yOffset = 10;
    public static int  squareSize = 86;
    public Game1()
    {
        _graphics = new GraphicsDeviceManager(this);
        Content.RootDirectory = "Content";
        IsMouseVisible = true;
        _graphics.PreferredBackBufferWidth = 720;
        _graphics.PreferredBackBufferHeight = 720;
        _graphics.ApplyChanges();
    }

    protected override void Initialize()
    {

        currentlyHeldPiece = null;

        pToMove = true;
        wCastleK = "K";
        wCastleQ = "Q";
        bCastleK = "k";
        bCastleQ = "q";
        enPass = " - ";

        squareLetter = new char[] 
        {
            'a',
            'b',
            'c',
            'd',
            'e',
            'f',
            'g',
            'h'
        };

   
        fullMoves = 0;


        bockFish = new Process();
        bockFish.StartInfo.FileName = "BockFish.exe";
        bockFish.StartInfo.UseShellExecute = false;
        bockFish.StartInfo.RedirectStandardInput = true;
        bockFish.StartInfo.RedirectStandardOutput = true;
        bockFish.Start();
        AppDomain.CurrentDomain.ProcessExit += closeBockFish;
    
      
/*
        sendCommand("uci");
        Console.WriteLine("Received: " + readAnswer());
        Console.WriteLine("Received: " + readAnswer());
*/ 
        sendCommand("ucinewgame");
        sendCommand("position startpos");
       // sendCommand("position fen rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1");
     //   sendCommand("position fen r7/3kq3/7P/6p1/1K6/8/8/8 w - - 0 1");
        
        sendCommand("isready");
        //sendCommand("position 5r2/pp5p/5n1k/2pp4/3n1R2/2PP3b/PP2rP1N/R4K2 w - - 3 30");
        readAnswer();

        legalMoves = getLegalMoves();
        
        base.Initialize();
    }
    static void closeBockFish(object sender, EventArgs e){
        Console.WriteLine("Closing bockfish");
        bockFish.StandardInput.WriteLine("quit");
        bockFish.StandardInput.Close();
        bockFish.Close();
    }

    static void sendCommand(string command){
        Console.WriteLine("Sent: "+ command);
        bockFish.StandardInput.WriteLine(command);
    }

    static string readAnswer(){
        string answer = bockFish.StandardOutput.ReadLine();
        Console.WriteLine("Recived:" + answer);
        return answer;
    }

    int[,] getLegalMoves(){
        sendCommand("getfullegal");
        //sendCommand("getlegal");
        string answer = readAnswer();
        string[] moves = answer.Split(" ");
        

       int[,] newLegal = new int[64,28];

        for (int i = 0; i < moves.Length-1; i++)
        {
            int SourceIndex = Int32.Parse(Char.ToString(moves[i][0])) + Int32.Parse(Char.ToString(moves[i][1])) *8 ;
            int TargetIndex = Int32.Parse(Char.ToString(moves[i][2]))  + Int32.Parse(Char.ToString(moves[i][3])) *8 ;
            for (int i2 = 0; i2 < 28; i2++)
            {
                if (newLegal[SourceIndex, i2] == 0) {
                    
                    newLegal[SourceIndex, i2] = TargetIndex+1;
                    break;
                }
            }
        }

        return newLegal;
    }

    void sendMoveToBock(Vector2 origin, Vector2 target){

        string SourceX = ((int)origin.X).ToString();
        string SourceY = ((int)origin.Y).ToString();
        string TargetX = ((int)target.X).ToString();
        string TargetY = ((int)target.Y).ToString();

        sendCommand("play " + SourceX + SourceY + TargetX + TargetY);
        Console.WriteLine("SentPlay: "+ readAnswer() ) ;
    }

    void getNewPos(){
        sendCommand("getpos");
        string answer = readAnswer();
        string[] pie = answer.Split(" ");
        
        string[] squareL = new string[] 
        {
            "a",
            "b",
            "c",
            "d",
            "e",
            "f",
            "g",
            "h"
        };

        for (int i = 1; i < pie.Length-1; i++)
        {
          //  Console.WriteLine(pie[i]);
            int sqareX = Int32.Parse(Char.ToString(pie[i][0]));
            int sqareY = Int32.Parse(Char.ToString(pie[i][1]));
            string col = Char.ToString(pie[i][2]);
            
            int type = Int32.Parse(Char.ToString(pie[i][3]));

            if (type == 0){
                boardArray[sqareY][sqareX] = null;
                continue;
            }

            string ty = squareL[type];

            if (col == "w"){
                ty = ty.ToUpper();
            }

           // boardArray[sqareY][sqareX].type = ty;
           // boardArray[sqareY][sqareX].c = col;
            Texture2D tex;

            if (col == "w"){
                tex = wTyToText[type];
            }else{
                tex = bTyToText[type];
            }
            //boardArray[sqareY][sqareX].sprite = tex;

            boardArray[sqareY][sqareX] = new Piece(ty, tex, col);

        }
    }

    protected override void LoadContent()
    {
        _spriteBatch = new SpriteBatch(GraphicsDevice);
        gw = Window;

        // TODO: use this.Content to load your game content here
        backgroundTexture = Content.Load<Texture2D>("chessboard");
        ballTexture = Content.Load<Texture2D>("ball");
        pawnTextureB = Content.Load<Texture2D>("black-pawn");
        pawnTextureW = Content.Load<Texture2D>("white-pawn");
        rookTextureB = Content.Load<Texture2D>("black-rook");
        rookTextureW = Content.Load<Texture2D>("white-rook");
        knightTextureB = Content.Load<Texture2D>("black-knight");
        knightTextureW = Content.Load<Texture2D>("white-knight");
        bishopTextureB = Content.Load<Texture2D>("black-bishop");
        bishopTextureW = Content.Load<Texture2D>("white-bishop");
        queenTextureB = Content.Load<Texture2D>("black-queen");
        queenTextureW = Content.Load<Texture2D>("white-queen");
        kingTextureB = Content.Load<Texture2D>("black-king");
        kingTextureW = Content.Load<Texture2D>("white-king");
        dotTexture = Content.Load<Texture2D>("dot");
        currentFont = Content.Load<SpriteFont>("font");

        boardArray = new Piece[][] {
          new Piece[]{
            new Piece("r", rookTextureB, "b"), 
            new Piece("n", knightTextureB, "b"), 
            new Piece("b", bishopTextureB, "b"),
            new Piece("q", queenTextureB, "b"),
            new Piece("k", kingTextureB, "b"),
            new Piece("b", bishopTextureB, "b"),
            new Piece("n", knightTextureB, "b"),
            new Piece("r", rookTextureB, "b")},
            new Piece[]{
               new Piece("p", pawnTextureB, "b"), 
               new Piece("p", pawnTextureB, "b"), 
               new Piece("p", pawnTextureB, "b"), 
               new Piece("p", pawnTextureB, "b"), 
               new Piece("p", pawnTextureB, "b"), 
               new Piece("p", pawnTextureB, "b"), 
               new Piece("p", pawnTextureB, "b"), 
               new Piece("p", pawnTextureB, "b")},
            new Piece[]{ null, null, null, null, null, null, null, null },
            new Piece[]{ null, null, null, null, null, null, null, null },
            new Piece[]{ null, null, null, null, null, null, null, null },
            new Piece[]{ null, null, null, null, null, null, null, null },
            new Piece[]{
                new Piece("P", pawnTextureW, "w"), 
                new Piece("P", pawnTextureW, "w"), 
                new Piece("P", pawnTextureW, "w"), 
                new Piece("P", pawnTextureW, "w"), 
                new Piece("P", pawnTextureW, "w"), 
                new Piece("P", pawnTextureW, "w"), 
                new Piece("P", pawnTextureW, "w"), 
                new Piece("P", pawnTextureW, "w")},
          new Piece[]{
            new Piece("R", rookTextureW, "w"),
            new Piece("N", knightTextureW, "w"),
            new Piece("B", bishopTextureW, "w"),
            new Piece("Q", queenTextureW, "w"),
            new Piece("K", kingTextureW, "w"),
            new Piece("B", bishopTextureW, "w"),
            new Piece("N", knightTextureW, "w"),
            new Piece("R", rookTextureW, "w")
            }};

        wTyToText = new Texture2D[]{
            pawnTextureW,
            pawnTextureW,
            knightTextureW,
            bishopTextureW,
            rookTextureW,
            queenTextureW,
            kingTextureW,
        };

        bTyToText = new Texture2D[]{
            pawnTextureB,
            pawnTextureB,
            knightTextureB,
            bishopTextureB,
            rookTextureB,
            queenTextureB,
            kingTextureB,
        };

        getNewPos();
    }

    bool pressed = false;

    
    protected override void Update(GameTime gameTime)
    {
        if (GamePad.GetState(PlayerIndex.One).Buttons.Back == ButtonState.Pressed || Keyboard.GetState().IsKeyDown(Keys.Escape))
            Exit();

        // TODO: Add your update logic here

        var mstate = Mouse.GetState();
        var kstate = Keyboard.GetState();
      
        var isClicked = mstate.LeftButton == ButtonState.Pressed;
        if (isClicked)
        {
            //System.Console.WriteLine("X: " + mstate.X + "  Y: "+ mstate.Y);
            currentlyHeldPiece = findClosestPiece(mstate);
            try
            {
                currentlyHeldPiece.pos.X = mstate.X;
                currentlyHeldPiece.pos.Y = mstate.Y; 
            }
            catch (System.Exception)
            {
                 
            }
        } else if (mstate.LeftButton == ButtonState.Released && currentlyHeldPiece != null)
        {
            dropPiece(currentlyHeldPiece, mstate);
        }


        if (kstate.IsKeyUp(Keys.U) && kstate.IsKeyUp(Keys.P)){
            pressed = false;
           
        }

        if (kstate.IsKeyDown(Keys.U)&& !pressed){
            pressed = true;
            Console.WriteLine("U");
            sendCommand("undo");
            Console.WriteLine(readAnswer());

            getNewPos();
            legalMoves =  getLegalMoves();
        }

         if (kstate.IsKeyDown(Keys.P)&& !pressed){
            pressed = true;
            Console.WriteLine("U");
            sendCommand("plabest");
            Console.WriteLine(readAnswer());

            getNewPos();
            legalMoves =  getLegalMoves();
        }
        base.Update(gameTime);
    }

    protected override void Draw(GameTime gameTime)
    {
        GraphicsDevice.Clear(Color.CornflowerBlue);

        // TODO: Add your drawing code here
        _spriteBatch.Begin();
        _spriteBatch.Draw(backgroundTexture, new Vector2(0, 0), Color.White); //Board
        drawBoard();
        if (currentlyHeldPiece != null)
        {
            drawFloatingPiece(currentlyHeldPiece);
        }
        _spriteBatch.End();

        base.Draw(gameTime);
    }

    void drawBoard() 
    {
        for (int y = 0; y < boardArray.Length; y++)
        {
            for (int x = 0; x < boardArray[y].Length; x++)
            {
                Piece piece = boardArray[y][x];
                if (piece != null)
                {

                    _spriteBatch.Draw(piece.sprite, 
                    new Vector2(squareSize*x + xOffset + squareSize/2, squareSize*y + yOffset + squareSize/2 ),
                    null, Color.White, 0f, new Vector2(piece.sprite.Width/2,piece.sprite.Height/2), Vector2.One, SpriteEffects.None, 0f);
                
                }
            }
        }
    }

    FloatingPiece findClosestPiece(MouseState mstate) 
    {
        if (currentlyHeldPiece == null) 
        {
          
            float x = System.MathF.Floor( (mstate.X - xOffset) / squareSize ); 
            float y = System.MathF.Floor( (mstate.Y - yOffset) / squareSize ) ;
            try{
                Piece closestPiece = boardArray[(int)y][(int)x];
                Vector2 origin = new Vector2(x,y);

                if (closestPiece != null) 
                {
                    boardArray[(int)y][(int)x] = null;
                    return new FloatingPiece(closestPiece.type, closestPiece.sprite, closestPiece.c, origin, origin);
                }
            }
            catch (System.Exception) {return null;}
        }
        return currentlyHeldPiece;
    }

    void dropPiece(FloatingPiece heldPiece, MouseState mstate)
    {
        Vector2 origin = heldPiece.origin;
        float x = System.MathF.Floor( (mstate.X - xOffset) / squareSize ) ;
        float y = System.MathF.Floor( (mstate.Y - yOffset) / squareSize ) ;
        bool validMove = false;
        try{
            for (int i = 0; i < 28; i++) 
            {
                if (legalMoves[(int)origin.Y*8 + (int)origin.X, i  ] == y*8+x+1)
                {
                    validMove = true;
                }
            }
        }
        catch (System.Exception){}
        
        if (validMove)
        {
          Vector2 target = new Vector2(x,y); 

          updateSpecialMoves(heldPiece, x, y);
         
          if (heldPiece.c == "b") fullMoves++;
          if (heldPiece.type.ToUpper() != "P" && boardArray[(int)y][(int)x] == null) halfMoves++; else halfMoves = 0;
          boardArray[(int)y][(int)x] = new Piece(heldPiece.type, heldPiece.sprite, heldPiece.c);
          //boardArray[(int)origin.Y][(int)origin.X] = new Piece(heldPiece.type, heldPiece.sprite, heldPiece.c);
          currentlyHeldPiece = null;
          pToMove = !pToMove;
          Console.WriteLine(convertPosToFen());

          
          sendMoveToBock(origin, target);
          legalMoves = getLegalMoves();
          getNewPos();
        } else 
        {
            currentlyHeldPiece = null;
            boardArray[(int)origin.Y][(int)origin.X] = new Piece(heldPiece.type, heldPiece.sprite, heldPiece.c);
        }
    }

    void drawFloatingPiece(FloatingPiece piece)
    {
        Vector2 origin = piece.origin;
          try 
          {
            for (int i = 0; i < 28; i++)
            {
                if (legalMoves[(int)origin.Y*8 + (int)origin.X, i  ] != 0){
                    int x = (legalMoves[(int)origin.Y*8 + (int)origin.X, i] - 1) % 8;
                    int y = (legalMoves[(int)origin.Y*8 + (int)origin.X, i] - 1) / 8;

                    _spriteBatch.Draw(
                      dotTexture,
                      new Vector2(squareSize*x + xOffset + squareSize/2, squareSize*y + yOffset + squareSize/2 ),
                      null,
                      Color.White,
                      0f,
                      new Vector2(dotTexture.Width / 2, dotTexture.Height / 2),
                      Vector2.One,
                      SpriteEffects.None,
                      0f
                      );


                 
                }
              }
          }
          catch (System.Exception) {}
        
        _spriteBatch.Draw(
            piece.sprite,
            piece.pos,
            null,
            Color.White,
            0f,
            new Vector2(piece.sprite.Width / 2, piece.sprite.Height / 2),
            Vector2.One,
            SpriteEffects.None,
            0f
            );
    }

    void updateSpecialMoves(FloatingPiece heldPiece, float targetX, float targetY){
         switch (heldPiece.type)
          {
            case "K":
              wCastleK = wCastleQ = "";
              break;
            case "k":
              bCastleK = bCastleQ = "";
              break;

            case "R":
              if (heldPiece.origin.Y == 0)
              {
                  if (heldPiece.origin.X == 7)
                  {
                      wCastleK = "";
                  } else if (heldPiece.origin.X == 0) 
                  {
                      wCastleQ = "";
                  }
              }
              break;
            case "r":
              if (heldPiece.origin.Y == 7)
              {
                  if (heldPiece.origin.X == 7)
                  {
                      bCastleK = "";
                  } else if (heldPiece.origin.X == 0) 
                  {
                      bCastleQ = "";
                  }
              }
              break;

            case "P":
              if (targetY - heldPiece.origin.Y == -2) 
              {
                  enPass = " " + squareLetter[(int)targetX] + "3 ";
              }
              break;
            case "p":
              if (targetY - heldPiece.origin.Y == 2) 
              {
                  enPass = " " + squareLetter[(int)targetX] + "6 ";
              }
              break;
            default:
              break;
          }
    }

    string convertPosToFen()
    {
        string fen = "";
        int numOfEmpty;
        for (int y = 0; y < boardArray.Length; y++)
        {
            numOfEmpty = 0;
            for (int x = 0; x < boardArray[y].Length; x++)
            {
                Piece currentPiece = boardArray[y][x];
                if (currentPiece == null) 
                {
                    numOfEmpty++;
                } else
                {
                    if (numOfEmpty != 0) 
                    {
                        fen += numOfEmpty.ToString();
                        numOfEmpty = 0;
                    }
                    fen += currentPiece.type;
                }
            }
            if (numOfEmpty != 0) fen += numOfEmpty.ToString();
            fen += "/";
        }

        fen = fen.Remove(fen.Length - 1);
        fen += " ";
        if (pToMove) fen += "w "; else fen += "b ";

        fen += wCastleK;
        fen += wCastleQ;
        fen += bCastleK;
        fen += bCastleQ;
        if (wCastleK == "" && wCastleQ == "" && bCastleK == "" && bCastleQ == "") fen += "-";
        
        fen += enPass;
        enPass = " - ";

        fen += halfMoves.ToString() + " ";
        fen += fullMoves.ToString();

        return fen;
    }
}

class Piece
{
    public string type;
    public Texture2D sprite;
    public string c;
    public Piece(string _type, Texture2D texture, string color) 
    {
        type = _type;
        sprite = texture;
        c = color;
    }
}

class FloatingPiece : Piece
{
    public Vector2 pos;
    public Vector2 origin;
    public FloatingPiece(string type, Texture2D texture, string color, Vector2 position, Vector2 or) : base(type, texture, color) {
        pos = position;
        origin = or;
    }
}
